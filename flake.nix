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
        pkgs = import nixpkgs { inherit system; };
        rev = if builtins.hasAttr "shortRev" self then self.shortRev else self.dirtyShortRev;

        # Anytime dependencies update or change, we will need to update this.
        # This ensures a package is reproducible.
        vendorHash = "sha256-fHmVz1JTwKb6p6bZhP3Qt7LFeMIYi5uAOiDe1GWtoVw=";

        polyform =
          with pkgs;
          buildGoModule {
            inherit vendorHash;
            name = "polyform";
            src = ./.;
            env = {
              CGO_ENABLED = 0;
            };
            doCheck = false;
            subPackages = [ "cmd/polyform" ];
          };

        polywasm =
          with pkgs;
          buildGoModule {
            inherit vendorHash;
            name = "polywasm";
            src = ./.;
            env = {
              CGO_ENABLED = 0;
            };
            doCheck = false;
            subPackages = [ "cmd/polywasm" ];
          };

        examples =
          with pkgs;
          lib.mapAttrs (
            name: path:
            (buildGoModule {
              inherit name vendorHash;
              src = ./.;
              env = {
                CGO_ENABLED = 0;
              };
              subPackages = [ "examples/${name}" ];
            })
          ) (lib.filterAttrs (n: v: v == "directory") (builtins.readDir ./examples));

        pages =
          with pkgs;
          stdenvNoCC.mkDerivation {
            name = "pages";
            src = ./.;
            phases = [
              "unpackPhase"
              "buildPhase"
            ];
            nativeBuildInputs = [ go ];
            buildPhase = ''
              export HOME=$(mktemp -d)
              export DIST=$(mktemp -d)
              mkdir -p $out $GOCACHE

              cp -r ${polyform.goModules} ./vendor
              GOOS=js GOARCH=wasm go build -mod=vendor -o ./main.wasm ./cmd/polyform

              ${polywasm}/bin/polywasm build --version ${rev} --wasm ./main.wasm -o $DIST

              cp -r $DIST/* $out
            '';
          };

        mkExamplesCross =
          GOOS: GOARCH:
          with pkgs;
          lib.mapAttrs (
            name: path:
            (buildGoModule {
              inherit name vendorHash;
              src = ./.;
              env = {
                CGO_ENABLED = 0;
              };
              subPackages = [ "examples/${name}" ];
            }).overrideAttrs
              (old: old // { inherit GOOS GOARCH; })
          ) (lib.filterAttrs (n: v: v == "directory") (builtins.readDir ./examples));

        supportedGoPlatforms = [
          "linux/amd64"
          "linux/arm64"
          "darwin/arm64"
          "darwin/amd64"
          "windows/arm64"
          "windows/amd64"
        ];

        withEachPlatform =
          module:
          with pkgs;
          (lib.foldl lib.mergeAttrs { } (
            map (
              platform:
              let
                parts = lib.splitString "/" platform;
                GOOS = builtins.elemAt parts 0;
                GOARCH = builtins.elemAt parts 1;
              in
              {
                ${GOOS} = {
                  ${GOARCH} = module GOOS GOARCH;
                };
              }
            ) supportedGoPlatforms
          ));
      in
      {
        packages = {
          inherit
            polyform
            polywasm
            examples
            pages
            ;
          default = polyform;
          examplesCross = withEachPlatform mkExamplesCross;
        };

        apps = {
          release = {
            type = "app";
            program = toString (
              pkgs.writers.writeBash "release" ''
                ${pkgs.goreleaser}/bin/goreleaser release
              ''
            );
          };
        };

        devShell = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
            gotools
            go-tools
            goreleaser
          ];
        };
      }
    );
}
