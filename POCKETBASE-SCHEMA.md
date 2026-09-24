# PocketBase schema MVP

Se mantienen las 17 colecciones propuestas. `users` debe ser una colección Auth; el resto son Base collections. No se agregan colecciones: una `attempts` guarda tanto respuesta como evaluación; `mastery` es el estado agregado; `sessions/messages` separan continuidad conversacional de contenido pedagógico.

## Colecciones y función
- `users`: identidad/auth de PocketBase.
- `learning_goals`: usuario, dominio, propósito, nivel, minutos y estado.
- `learning_paths`: versión generada para un objetivo.
- `topics`: agrupación de conceptos por dominio.
- `concepts`: nodos enseñables del mapa.
- `concept_dependencies`: aristas prerequisite (`from`, `to`), únicas.
- `lessons`: contenido breve y comprobación asociada.
- `exercises`: prompt, tipo, respuesta esperada y rúbrica.
- `assessments`: evaluación estructurada, proveedor y modelo, sin secretos.
- `attempts`: respuesta del usuario y vínculo a assessment.
- `mastery`: dominio por usuario/concepto.
- `review_schedule`: próxima revisión, pospuesta a post-MVP salvo lectura.
- `sources`: fuente y cita, para investigación trazable.
- `sessions`: continuidad de objetivo.
- `messages`: texto de sesiones; no sustituye ejercicios.
- `media_assets`: archivos futuros de lección/voz.
- `pronunciation_attempts`: capa de voz futura, no bloquea MVP.

## Reglas
Relations siempre apuntan a IDs; ownership se valida en backend. Campos de proveedor solo contienen metadatos. Nunca guardar API keys, tokens, prompts con secretos ni credenciales.

Ver `infra/pocketbase-schema.json` para el export declarativo inicial.
