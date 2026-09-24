# Voice provider contract

Estado: preparado, no implementado en el MVP.

Puertos futuros: `SpeechToText(audio) -> transcript, confidence`; `TextToSpeech(text, voice) -> media asset reference`; `PronunciationEvaluate(audio, target) -> score, feedback`.

Audio se almacenará como `media_assets`; intentos como `pronunciation_attempts`. Las credenciales vivirán en variables/secret manager del backend, nunca en PocketBase ni frontend. Proveedor, formatos, retención y consentimiento quedan pendientes.
