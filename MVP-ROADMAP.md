# MVP roadmap

## Fase 1 — vertical de interfaz (completada)
Objetivo → mapa de 3 conceptos → lección → ejercicio abierto → evaluación heurística mock → mastery/progreso. La interfaz web y el flujo fueron verificados con pruebas Go y un recorrido de navegador en local.

## Fase 2 — acceso y persistencia (siguiente)
PocketBase Auth y Base collections; propiedad de datos y reglas de acceso; sesiones/mensajes; reanudación del objetivo. No exponer PocketBase ni credenciales administrativas al navegador o a Internet.

## Fase 3 — aprendizaje generado y evaluado
Integrar el proveedor OpenCode detrás del contrato confirmado. No inventar endpoints ni modelos; mantener el mock etiquetado hasta configurar una credencial segura. Añadir contenidos específicos y fuentes trazables donde aplique.

## Fase 4 — práctica adaptativa
Quizzes variados, repetición espaciada, adaptación al desempeño y revisión de calidad del contenido.

## Fuera del MVP
Voz, cursos masivos, comunidad, marketplace y analítica avanzada.

## Gate de despliegue
La vertical local funciona. Se permite publicar un preview de prueba autorizado, claramente etiquetado y sin datos sensibles; no equivale a producción. Antes de invitar a más personas o guardar datos reales, cerrar Fase 2 (autenticación, PocketBase, persistencia y aislamiento). El dominio de prueba se configura solo tras comprobar DNS y routing de Dokploy.
