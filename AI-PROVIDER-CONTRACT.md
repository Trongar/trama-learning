# AI provider contract

## Estado
Pendiente confirmar el contrato concreto de OpenCode. No se inventan URL, método ni modelo.

## Puerto interno
`Evaluator.Evaluate(ctx, EvaluationInput) -> EvaluationOutput`.

Entrada: dominio, concepto, prompt, respuesta, respuesta esperada, rúbrica.
Salida mínima: `correct` boolean, `score` 0..1, `explanation`, `error_tags[]`, `facts[]`, `inferences[]`, `pedagogy[]`, `provider`, `model`.

## Seguridad
Solo backend; timeouts, límites de tamaño y validación estricta. El frontend nunca recibe URL/key del proveedor. Logs no incluyen secretos ni respuesta completa si contiene datos sensibles.

## MVP
`MockEvaluator` compara una respuesta normalizada para pruebas reproducibles y se etiqueta como `mock`. El adaptador real se implementará solo cuando exista documentación/credencial configurada.
