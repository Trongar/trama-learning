# Trama · prototipo de aprendizaje

Trama es una experiencia web para convertir una meta de aprendizaje en un mapa de conceptos, lecciones breves, práctica y seguimiento del progreso. El repositorio conserva el nombre interno `edg-learning-mvp`.

## Ejecutar localmente

Requiere Go 1.24 o posterior:

```sh
go run .
```

Abre `http://localhost:8080`.

## Verificar

```sh
go test -race ./...
go vet ./...
go build -o trama .
```

## Recorrido disponible

1. Crear una meta con tema, nivel, propósito y tiempo semanal.
2. Ver un mapa inicial de tres pasos.
3. Abrir una lección de demostración.
4. Explicar una idea con palabras propias.
5. Recibir retroalimentación y ver una señal de progreso.

Los formularios funcionan sin JavaScript; HTMX mejora la navegación cuando está disponible.

## Límites actuales

- La ruta y las lecciones son demostrativas: aún no se generan contenidos específicos con IA.
- La evaluación usa una heurística local etiquetada `mock`, no la API de OpenCode.
- Los objetivos y el progreso solo viven en memoria; se pierden al reiniciar el proceso.
- Aún no hay cuentas, persistencia de PocketBase ni sincronización entre dispositivos.
- La capa de voz todavía no está conectada.

Por ahora es una prueba funcional de interfaz y flujo, no un servicio listo para producción. Consulta `MVP-ROADMAP.md` para las siguientes fases.
