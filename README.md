# WOL Dashboard

Panel web para registrar dispositivos, comprobar si responden por TCP y enviar paquetes Wake-on-LAN.

## Desarrollo

```bash
pnpm install
docker compose up --build
```

El backend usa SQLite persistente en el volumen `wol_data`. Usa `network_mode:
host` para poder enviar el broadcast WOL a la LAN; el frontend accede a la API
mediante `host.docker.internal`.

La aplicación es web y no incluye cliente móvil.
