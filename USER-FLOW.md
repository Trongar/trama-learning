# User flow prioritario

1. Usuario crea cuenta (PocketBase Auth; tracer bullet usa `user_id` demo).
2. Completa objetivo con dominio, nivel actual, propósito y minutos disponibles.
3. Backend genera una ruta pequeña y mapa de 3 conceptos con dependencias.
4. Usuario abre el primer concepto y ve una lección breve.
5. La lección muestra una comprobación; responde texto.
6. Backend evalúa mediante el puerto de proveedor (mock identificado en MVP).
7. Se muestra resultado, explicación del error y siguiente acción.
8. Backend actualiza `mastery`; el mapa y progreso reflejan el cambio.
9. `sessions` conserva la continuidad para reanudar objetivo.

Criterio E2E: tras el POST de intento, `mastery` del concepto y `progress` del objetivo cambian y la respuesta incluye evaluación explicada.
