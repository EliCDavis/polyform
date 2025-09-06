{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      utils,
      ...
    }:
    utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };

        vendorHash = builtins.readFile ./go.mod.sri;
        rev = if builtins.hasAttr "shortRev" self then self.shortRev else self.dirtyShortRev;
        src = builtins.path {
          path = ./.;
          name = "source";
          filter =
            path: type:
            !(builtins.elem (baseNameOf path) [
              "flake.nix"
              "flake.lock"
            ]);
        };

        supportedGoPlatforms = [
          "linux/amd64"
          "linux/arm64"
          "darwin/arm64"
          "darwin/amd64"
          "windows/arm64"
          "windows/amd64"
        ];

        cmd =
          with pkgs;
          lib.mapAttrs (
            name: type:
            buildGoModule {
              inherit name vendorHash src;
              env = {
                CGO_ENABLED = 0;
              };
              subPackages = [
                "cmd/${name}"
              ];
              preBuild = ''
                cp -r ${website}/* ./generator/edit/html/
              '';
            }
          ) (lib.filterAttrs (n: v: v == "directory") (builtins.readDir ./cmd));

        examples =
          with pkgs;
          lib.mapAttrs (
            name: type:
            buildGoModule {
              inherit name vendorHash src;
              env = {
                CGO_ENABLED = 0;
              };
              subPackages = [
                "examples/${name}"
              ];
              preBuild = ''
                cp -r ${website}/* ./generator/edit/html/
              '';
            }
          ) (lib.filterAttrs (n: v: v == "directory") (builtins.readDir ./examples));

        release =
          with pkgs;
          stdenvNoCC.mkDerivation {
            inherit src;
            name = "release";
            phases = [
              "unpackPhase"
              "buildPhase"
            ];
            nativeBuildInputs = [
              zip
            ];
            buildPhase = ''
              export HOME=$(mktemp -d)
              export BUILD=$(mktemp -d)
              shopt -s globstar
              mkdir -p $out

              ln -s ${cmd.polyform.goModules} ./vendor

              cp -r ${website}/* ./generator/edit/html/

              for platform in ${builtins.toString supportedGoPlatforms}; do
                GOOS=$(echo $platform | cut -d'/' -f1)
                GOARCH=$(echo $platform | cut -d'/' -f2)
                ARTIFACT_NAME=$(printf "programs_%s_%s" $GOOS $GOARCH)

                mkdir "$BUILD/$ARTIFACT_NAME"
                GOOS=$GOOS GOARCH=$GOARCH ${go}/bin/go build -o "$BUILD/$ARTIFACT_NAME" --mod=vendor "./cmd/..." "./examples/..." &
              done
              wait

              for platform in ${builtins.toString supportedGoPlatforms}; do
                GOOS=$(echo $platform | cut -d'/' -f1)
                GOARCH=$(echo $platform | cut -d'/' -f2)
                ARTIFACT_NAME=$(printf "programs_%s_%s" $GOOS $GOARCH)

                if [ "$GOOS" = "windows" ]; then
                  find "$BUILD/$ARTIFACT_NAME" -type f -exec zip -j "$out/$ARTIFACT_NAME.zip" {} \; &
                else
                  tar -czf "$out/$ARTIFACT_NAME.tar.gz" -C "$BUILD/$ARTIFACT_NAME" . &
                fi
              done
              wait

              for f in $out/*.tar.gz $out/*.zip; do
                sha256sum "$f" >> $out/checksums.txt
              done
            '';
          };

        pages =
          with pkgs;
          stdenvNoCC.mkDerivation {
            inherit src;
            name = "pages";
            phases = [
              "unpackPhase"
              "buildPhase"
            ];
            nativeBuildInputs = [ go ];
            buildPhase = ''
              export HOME=$(mktemp -d)
              export DIST=$(mktemp -d)
              mkdir -p $out

              ln -s ${cmd.polyform.goModules} ./vendor

              cp -r ${website}/* ./generator/edit/html/

              GOOS=js GOARCH=wasm go build -mod=vendor -o ./main.wasm ./cmd/polyform

              ${cmd.polywasm}/bin/polywasm build --version ${rev} --wasm ./main.wasm -o $DIST

              cp -r $DIST/* $out
            '';
          };

        website =
          with pkgs;
          buildNpmPackage {
            inherit src;
            pname = "website";
            version = "0.0.1";
            npmBuildScript = "build-dev";
            npmDepsHash = builtins.readFile ./package-lock.json.sri;
            installPhase = ''
              runHook preInstallPhase

              mkdir -p $out
              cp -r ./generator/edit/html/* $out

              runHook postInstallPhase
            '';
          };

        apps = {
          # This can be run with:
          #   nix run .#sri-check-up
          #
          # It generates the sub-resource integrity hashes for both go and node dependencies.
          # This should be run anytime dependendencies change in this project and the results checked-in to vcs.
          # In a remote build environment, it prints the SRI hash difference into logs for easy update.
          sri-check-up = {
            type = "app";
            program = toString (
              pkgs.writeShellScript "update-sris" ''
                OUT=$(mktemp -d -t nar-hash-XXXXXX)
                rm -rf "$OUT"

                ${pkgs.go}/bin/go mod vendor -o "$OUT"
                GO_MOD_SRI=$(${pkgs.go}/bin/go run tailscale.com/cmd/nardump@v1.86.4 --sri "$OUT")
                rm -rf "$OUT"

                PACKAGE_LOCK_SRI=$(${pkgs.prefetch-npm-deps}/bin/prefetch-npm-deps ./package-lock.json)

                # Print the SRI diff in CI, otherwise update SRI if run locally
                if [ -n "$CI" ]; then
                   CHECK_GO_MOD_SRI=$(<go.mod.sri)
                   CHECK_PACKAGE_LOCK_SRI=$(<package-lock.json.sri)

                   if [ "$GO_MOD_SRI" != "$CHECK_GO_MOD_SRI" ]; then
                      echo "go.mod.sri mismatch"
                      echo "specified: $CHECK_GO_MOD_SRI"
                      echo "got: $GO_MOD_SRI"
                      echo "If this difference is expected, please replace the SRI hash in this file with the one we got"
                      exit 1
                   fi

                   if [ "$PACKAGE_LOCK_SRI" != "$CHECK_PACKAGE_LOCK_SRI" ]; then
                      echo "package-lock.json.sri mismatch"
                      echo "specified: $CHECK_PACKAGE_LOCK_SRI"
                      echo "got: $PACKAGE_LOCK_SRI"
                      echo "If this difference is expected, please replace the SRI hash in this file with the one we got"
                      exit 1
                   fi
                else
                   echo "go.mod.sri: Compute and store hash..."
                   echo "$GO_MOD_SRI" > ./go.mod.sri

                   echo "package-lock.json.sri: Compute and store hash..."
                   echo "$PACKAGE_LOCK_SRI" > ./package-lock.json.sri
                fi
              ''
            );
          };

          # Helpful script if Github Actions ever gits in a weird cache state. Clears all caches for the entire repo.
          # May increase subsequent build times until a full pipeline completes and saves a new cache.
          clear-gh-action-cache = {
            type = "app";
            program = toString (
              pkgs.writeShellScript "clear-gh-action-cache" ''
                ${pkgs.gh}/bin/gh cache delete --all --repo EliCDavis/polyform
              ''
            );
          };

          gh-release = {
            type = "app";
            program = toString (
              pkgs.writeShellScript "gh-release" ''
                ${pkgs.gh}/bin/gh release upload $GITHUB_REF_NAME ${release}/* --clobber
              ''
            );
          };
        };
      in
      {
        inherit apps;

        packages = {
          inherit pages release website;
        }
        // cmd
        // examples;

        devShells = {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go
              gopls
              gotools
              go-tools
            ];
          };
        };
      }
    );
}
