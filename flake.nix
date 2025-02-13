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

        # Anytime dependencies update or change, this should be updated.
        # This ensures a package is reproducible.
        vendorHash = "sha256-fHmVz1JTwKb6p6bZhP3Qt7LFeMIYi5uAOiDe1GWtoVw=";
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
              inherit name vendorHash;
              src = ./.;
              env = {
                CGO_ENABLED = 0;
              };
              subPackages = [
                "cmd/${name}"
              ];
            }
          ) (lib.filterAttrs (n: v: v == "directory") (builtins.readDir ./cmd));

        examples =
          with pkgs;
          lib.mapAttrs (
            name: type:
            buildGoModule {
              inherit name vendorHash;
              src = ./.;
              env = {
                CGO_ENABLED = 0;
              };
              subPackages = [
                "examples/${name}"
              ];
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
              GOOS=js GOARCH=wasm go build -mod=vendor -o ./main.wasm ./cmd/polyform

              ${cmd.polywasm}/bin/polywasm build --version ${rev} --wasm ./main.wasm -o $DIST

              cp -r $DIST/* $out
            '';
          };
      in
      {
        packages =
          {
            inherit pages release;
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
