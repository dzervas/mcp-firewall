{
  outputs =
    { nixpkgs, ... }:
    let
      supportedSystems = [
        "x86_64-linux"
        "x86_64-darwin"
        "aarch64-linux"
        "aarch64-darwin"
      ];

      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in
    rec {
      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};
        in
        rec {
          mcp-firewall = pkgs.buildGoModule rec {
            pname = "mcp-firewall";
            version = "0.1.0";
            src = ./.;
            vendorHash = "sha256-vvsSHF7UiPtxjCXsnpD8D2rVqQNakfo0Oqw7wPVsYCM=";
            subPackages = [ "./cmd" ];
            # CGO_ENABLED = 0;
            ldflags = [
              "-extldflags"
              "-static"
              "-s -w"
              "-X main.builtBy=nix-flake"
              "-X main.Version=${version}"
            ];
            postInstall = ''
              mv "$out/bin/cmd" "$out/bin/mcp-firewall"
            '';
            meta.mainProgram = "mcp-firewall";
          };
          default = mcp-firewall;
        }
      );

      checks = forAllSystems (system: {
        mcp-firewall = packages.${system}.mcp-firewall.overrideAttrs (_: {
          doCheck = true;
          checkPhase = "go test ./...";
        });
      });

      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go
              gopls
              gotools
              jsonnet
            ];
          };
        }
      );
    };
}
