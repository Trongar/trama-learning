# Trama · MVP de aprendizaje

Trama convierte una meta de aprendizaje en una ruta breve de conceptos, práctica y seguimiento. El servidor activo está escrito en Go y PocketBase; cualquier archivo `edg_app.py` o `tests/test_vertical.py` que siga apareciendo en GitHub pertenece al prototipo Python anterior y no forma parte del build actual. El repositorio conserva el nombre interno `edg-learning-mvp`.

## Espacios personales para la prueba

Trama usa **PocketBase Auth**: cada persona crea su propia cuenta con su correo y contraseña. PocketBase persiste cada ruta y aplica reglas de propietario para que una cuenta no pueda listar, abrir, modificar ni borrar las rutas de otra. Cada cuenta nueva recibe su propia ruta inicial de tres pasos (“Aprender a aprender”); no se comparte una ruta de demostración global.

Comparte el enlace de Trama, no una cuenta común. Cada persona debe registrarse por separado. La autenticación usa tokens de PocketBase en el almacenamiento de sesión de la pestaña; cerrar sesión la elimina. No hay verificación por correo ni recuperación de contraseña configuradas, así que cada tester debe recordar su contraseña. No uses información sensible en este prototipo.

## Ejecutar localmente

Requiere Go 1.27 o posterior:

```sh
TRAMA_DATA_DIR=./pb_data go run . serve --http=127.0.0.1:8080
```

Abre `http://127.0.0.1:8080`. PocketBase crea y migra sus colecciones al arrancar. `pb_data/` contiene la base SQLite local; no la subas al repositorio.

## Verificar

```sh
go test -race -count=1 ./...
go vet ./...
go build -o trama .
```

Con Trama escuchando en `127.0.0.1:8091`, puedes ejercitar las reglas reales de PocketBase con dos cuentas de prueba locales:

```sh
python3 scripts/test-pocketbase-isolation.py
```

El script solo acepta URLs localhost, genera credenciales temporales en memoria y no las imprime. Verifica registro, ruta inicial individual, lectura por ID, intentos, aislamiento en listas y rechazo de transferencia de propietario.

## Recorrido

1. Crear una cuenta individual o iniciar sesión.
2. Recibir la ruta inicial privada o crear una nueva meta.
3. Explorar tres pasos, enviar una respuesta y guardar el progreso.
4. Cerrar sesión en esta pestaña cuando termines.

Las rutas se sirven como HTML/CSS/JavaScript; PocketBase atiende autenticación, persistencia y reglas de acceso. El endpoint de intentos exige una cuenta autenticada y comprueba el propietario antes de actualizar el progreso.

## Despliegue

El contenedor ejecuta PocketBase en el puerto 8080 y usa `/data` para SQLite. El Dockerfile declara ese directorio como volumen y el proceso corre sin privilegios. Conserva el volumen `/data` al volver a desplegar; aún no hay backups automáticos configurados.

PocketBase 0.40.4 es anterior a 1.0 y su propia documentación advierte que no se recomienda para sistemas críticos sin seguir sus cambios y migraciones. Este despliegue es un piloto de feedback, no un servicio de producción.

## Límites

- La ruta inicial y la evaluación son demostrativas; la evaluación se etiqueta `mock · heuristic-demo-v1`, no usa IA.
- No hay verificación ni recuperación por correo, inicio OAuth, exportación de datos ni backups automáticos.
- Cada tester necesita su propia cuenta; compartir credenciales comparte también la información.
- La capa de voz todavía no está conectada.

Consulta `ARCHITECTURE.md`, `MVP-ROADMAP.md` y `RISKS.md` para los detalles y fases siguientes.
