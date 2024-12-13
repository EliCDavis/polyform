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
    }:
    utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };

        # Anytime dependencies update or change, we will need to update this.
        # This ensures a package is reproducible.
        vendorHash = "sha256-7V82ePT5So0dgoQfMUEncejO8gMjLoKRS1ixncSiSHA=";

        polywasm =
          GOOS: GOARCH:
          with pkgs;
          (buildGoModule {
            inherit vendorHash;
            name = "polywasm";
            src = ./.;
            CGO_ENABLED = 0;
            subPackages = [ "cmd/polywasm" ];

            # uncomment to skip running tests as part of build
            # doCheck = false;
          }).overrideAttrs
            (
              old:
              old
              // {
                inherit GOOS GOARCH;
                CGO_ENABLED = 0;
              }
            );

        examples =
          GOOS: GOARCH:
          with pkgs;
          lib.mapAttrs (
            name: path:
            (buildGoModule {
              inherit name vendorHash;
              src = ./.;
              CGO_ENABLED = 0;
              subPackages = [ "examples/${name}" ];

              # uncomment to skip running tests as part of build
              # doCheck = false;
            }).overrideAttrs
              (
                old:
                old
                // {
                  inherit GOOS GOARCH;
                }
              )
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
          default = polywasm;
          examples = examples;
          crossBuild = {
            polywasm = withEachPlatform polywasm;
            examples = withEachPlatform examples;
          };
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
