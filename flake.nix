{
  description = "smd - Secure My Directory CLI tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        
        smd = pkgs.buildGoModule {
          pname = "smd";
          version = "0.0.5";
          
          src = ./.;
          
          vendorHash = null;
          
          ldflags = [
            "-s"
            "-w"
          ];
          
          postInstall = ''
            mv $out/bin/src $out/bin/smd
          '';
          
          meta = with pkgs.lib; {
            description = "Secure My Directory - CLI tool for containerized development";
            homepage = "https://github.com/LukasLow/smd-cli";
            license = licenses.mit;
            mainProgram = "smd";
          };
        };
      in
      {
        packages = {
          default = smd;
          smd = smd;
        };

        apps.default = {
          type = "app";
          program = "${smd}/bin/smd";
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
            gotools
            podman
          ];
        };
      });
}
