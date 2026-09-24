# User flow prioritario

1. Persona abre el enlace y registra su cuenta individual con correo y contraseña.
2. PocketBase crea una ruta inicial privada de tres pasos para esa cuenta; si falta, la aplicación asegura su creación.
3. La persona ve sus rutas y crea otra meta con dominio, nivel, propósito y minutos disponibles.
4. PocketBase lista únicamente los registros cuyo `user` corresponde a la cuenta autenticada.
5. La persona abre un paso, lee el scaffold de demostración y envía una explicación.
6. El endpoint autenticado verifica el propietario antes de evaluar y guardar respuesta/progreso.
7. PocketBase persiste la actualización y sus reglas impiden que otras cuentas lean, editen o transfieran la ruta.
8. La persona cierra sesión en su pestaña; sus amistades se registran con sus propias cuentas.

Criterio E2E de privacidad: dos cuentas reciben rutas iniciales con IDs distintos; la segunda obtiene 404 al leer o enviar una respuesta a un ID de la primera, no ve el registro en su lista y no puede crear ni transferir una ruta para otro usuario.

La evaluación actual es un mock heurístico, no IA; las lecciones son scaffolds. No ingresar información sensible.
