{
  description = "ShengOS Daemon";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "aarch64-linux" "x86_64-linux" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
      packageFor = system:
        (import nixpkgs { inherit system; }).buildGoModule {
          pname = "shengos-daemon";
          version = builtins.elemAt (builtins.match "([^\n]*\n)*version: ([0-9]+\\.[0-9]+\\.[0-9]+)\n?" (builtins.readFile ./version.yaml)) 1;
          src = self;
          vendorHash = null;
          subPackages = [ "." ];
          ldflags = [ "-s" "-w" ];
          meta = {
            description = "Host daemon for ShengOS fleet state sync over MQTT";
            homepage = "https://github.com/jasonkwh/shengos-daemon";
            mainProgram = "shengosd";
          };
          postInstall = ''
            mv "$out/bin/shengos-daemon" "$out/bin/shengosd"
          '';
        };
    in
    {
      packages = forAllSystems (system: {
        default = packageFor system;
        shengos-daemon = packageFor system;
      });

      nixosModules.default = { config, lib, pkgs, ... }: {
        options.services.shengos-daemon = {
          enable = lib.mkEnableOption "ShengOS Daemon";
          package = lib.mkOption {
            type = lib.types.package;
            default = self.packages.${pkgs.system}.default;
            description = "Package providing the shengosd executable.";
          };
          configFile = lib.mkOption {
            type = lib.types.str;
            default = "${self}/config/config.yaml";
            description = "Path to the daemon YAML configuration file.";
          };
        };

        config = lib.mkIf config.services.shengos-daemon.enable {
          systemd.services.shengos-daemon = {
            description = "ShengOS Daemon";
            wantedBy = [ "multi-user.target" ];
            after = [ "network-online.target" ];
            wants = [ "network-online.target" ];
            serviceConfig = {
              ExecStart = "${config.services.shengos-daemon.package}/bin/shengosd --config=${config.services.shengos-daemon.configFile} serve";
              Restart = "on-failure";
              RestartSec = 5;
            };
          };
        };
      };
    };
}
