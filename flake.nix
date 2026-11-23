{
  description = "ShengOS Daemon";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";

  outputs = { self, nixpkgs }:
    let
      systems = [ "aarch64-linux" "x86_64-linux" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
      packageFor = system:
        let
          pkgs = import nixpkgs { inherit system; };
          buildGoModule = pkgs.buildGoModule.override { go = pkgs.go_1_27; };
        in buildGoModule {
          pname = "shengos-daemon";
          version = builtins.elemAt (builtins.match "([^\n]*\n)*version: ([0-9]+\\.[0-9]+\\.[0-9]+)\n?" (builtins.readFile ./version.yaml)) 1;
          src = self;
          vendorHash = "sha256-VY3XEXwTv94nfbNVs8Ji8QRToToLC8X+vFED44BJhTY=";
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
            default = self.packages.${pkgs.stdenv.hostPlatform.system}.default;
            description = "Package providing the shengosd executable.";
          };
          configFile = lib.mkOption {
            type = lib.types.str;
            default = "${self}/config/config.yaml";
            description = "Path to the daemon YAML configuration file (the defaults).";
          };
          environment = lib.mkOption {
            type = lib.types.attrsOf lib.types.str;
            default = { };
            example = { CLIPBOARD_ENABLED = "false"; MQTTCONN_HOST = "localhost"; };
            description = ''
              Environment variables that override configFile. A key is the
              upper-cased config field path joined with underscores, e.g.
              MQTTCONN_KEEPALIVE. Do not put secrets here: the unit lives in
              the world-readable Nix store; use environmentFile instead.
            '';
          };
          environmentFile = lib.mkOption {
            type = lib.types.nullOr lib.types.str;
            default = null;
            description = ''
              Optional file of KEY=value lines loaded into the service
              environment, for secrets such as MQTTCONN_USERNAME and
              MQTTCONN_PASSWORD. Not copied to the Nix store.
            '';
          };
          requireGraphicalSession = lib.mkOption {
            type = lib.types.bool;
            default = false;
            description = ''
              Bind the user service to graphical-session.target (needed
              for clipboard access). When disabled, it starts from
              default.target, suitable for headless hosts.
            '';
          };
          user = lib.mkOption {
            type = lib.types.str;
            default = "jasonkwh";
            description = ''
              User whose systemd user manager runs the daemon. The unit is
              skipped for every other user. The user needs linger enabled
              for the daemon to run without an active login session.
            '';
          };
        };

        config =
          let
            cfg = config.services.shengos-daemon;
            target =
              if cfg.requireGraphicalSession
              then "graphical-session.target"
              else "default.target";
          in
          lib.mkIf cfg.enable {
            systemd.user.services.shengos-daemon = {
              description = "ShengOS Daemon";
              wantedBy = [ target ];
              partOf = lib.optionals cfg.requireGraphicalSession [ target ];
              after = lib.optionals cfg.requireGraphicalSession [ target ];
              unitConfig.ConditionUser = cfg.user;
              inherit (cfg) environment;
              serviceConfig = {
                ExecStart = "${cfg.package}/bin/shengosd --config=${cfg.configFile} serve";
                Restart = "on-failure";
                RestartSec = 5;
              } // lib.optionalAttrs (cfg.environmentFile != null) {
                EnvironmentFile = cfg.environmentFile;
              };
            };
          };
      };
    };
}
