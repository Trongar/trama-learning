# Trama — Product brief

## Tesis
Trama ayuda a una persona a convertir un objetivo de aprendizaje en pasos pequeños de conceptos, lecciones y comprobaciones; el sistema adapta el siguiente paso a la evidencia de desempeño.

## Usuario inicial
Persona autodidacta que quiere aprender cualquier dominio y dispone de poco tiempo. El MVP no asume idiomas.

## Problema
Los cursos grandes mezclan consumo con comprensión. El usuario necesita saber qué estudiar ahora, comprobarlo y recibir una explicación accionable cuando falla.

## MVP verificable
Crear una cuenta individual en PocketBase; recibir una ruta inicial propia; crear otra meta; ver un mapa pequeño de tres pasos; abrir una lección y responder; obtener una evaluación heurística claramente marcada; guardar mastery/progreso con acceso restringido al propietario.

## Antiobjetivos
Chatbot genérico, generación de cursos completos, marketplace, voz operativa, recomendaciones externas no trazables y analítica avanzada.

## Métrica de primera versión
Porcentaje de testers que completan la vertical: cuenta creada → ruta abierta → ejercicio respondido → progreso guardado.

## Decisiones
- Una lección siempre tiene al menos una comprobación.
- PocketBase gestiona autenticación, SQLite, migraciones y reglas de acceso por propietario.
- Cada usuario recibe una ruta inicial independiente; compartir enlace no comparte las rutas ni sustituye la cuenta individual.
- El backend es el único que hablará con el proveedor de IA.
- Las respuestas del proveedor deben distinguir hechos, inferencias y explicación pedagógica.
- Investigación externa y fuentes se incorporan después, preservando referencias.

## Límites
La ruta inicial y la evaluación son demostrativas; no hay proveedor de IA, correo de recuperación, voz ni backup automático. Este despliegue es un piloto de feedback, no un sistema crítico.

## Pendientes
Validar el contrato real de OpenCode, configurar SMTP y política de recuperación, verificar backup/restore y elegir proveedor de voz; todos quedan fuera del primer feedback.
