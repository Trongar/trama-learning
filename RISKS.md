# Risks

- **Contrato OpenCode desconocido — alto:** bloquea integración real. Mitigación: puerto + mock explícito; pendiente documentado.
- **Persistencia inicial en memoria — medio:** reinicio pierde datos. Mitigación: reemplazo por PocketBase en Fase 2; no presentar como producción.
- **Evaluación automática incorrecta — alto:** una etiqueta errónea puede enseñar mal. Mitigación: explicación obligatoria, rúbrica, feedback y pruebas; revisión humana futura.
- **Generación de mapas grandes — medio:** reduce finalización. Mitigación: mapa inicial limitado a 3 nodos.
- **Datos sensibles en prompts — medio:** riesgo de exposición. Mitigación: backend-only, límites y redacción de logs.
- **Voz prematura — bajo:** dispersa alcance. Mitigación: contrato solamente.
