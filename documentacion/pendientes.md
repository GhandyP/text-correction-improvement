# Pendientes — text-correction-improvement

Estado de lo que queda abierto en este repositorio. Se actualiza cuando algo cambia: lo que se cierra
se registra en [`odd/tasks/text-correction-v1.md`](../odd/tasks/text-correction-v1.md) y se borra de
acá. Última revisión: 2026-10-13.

## Estado verificado

| Punto | Estado |
| --- | --- |
| Rama `main` | 51e9a8c, remoto == local |
| Tareas ODD (T0–T11) | todas cerradas, con commit propio |
| Verificación automática | `make check` (gofmt, vet, build, test) en verde — 80 funciones de test |
| Llamada real al modelo | **no** cubierta por CI (manual, requiere credencial) |
| `TODO` / `FIXME` en el árbol | ninguno |

## 1. Decisiones abiertas (requieren una persona)

1. **Commit del borrado del caso copiado.** La copia que originó este proyecto
   (`casos/09-text-correction-improvement` en el repositorio del cookbook) se borró del árbol de
   trabajo y quedó **sin commitear** en ese repositorio: no se commitea ahí sin autorización
   explícita. Se recupera con `git restore casos/09-text-correction-improvement` mientras siga así.
2. **Trazabilidad del hallazgo advisory `R4-001`.** Su texto literal no es recuperable del store local
   de review (solo sobrevivió el recibo de consumo terminal de la lineage). Queda registrada la
   ubicación que informó (`internal/cli/cli.go:168-169`), el defecto concreto que había ahí y su
   corrección. Si se necesita el texto original del revisor, no existe artefacto que lo contenga.

## 2. Qué no cubre la verificación automática

1. **La llamada real a Gemini.** CI corre todo el pipeline con un generador inyectado y respuestas
   grabadas; la conversación con el modelo se prueba a mano con `GOOGLE_API_KEY`.
2. **`gemini-flash-latest` es un alias móvil.** El comportamiento del modelo por defecto puede cambiar
   sin que cambie ningún commit. `ADK_MODEL` permite fijar uno concreto.
3. **Sin reintentos ni backoff.** Un error transitorio del modelo es un fallo del run
   (`model_call_failed`), no se reintenta.
4. **Sin límite de tamaño.** Ni de entrada ni de cantidad de cambios: un texto muy largo se envía tal
   cual.
5. **Sin timeouts propios.** El corte depende del `context` que pase quien invoca el binario.
6. **Solo reemplazos.** Un borrado o una inserción puros tienen que anclarse con contexto: ambos lados
   de un cambio deben ser localizables en el texto y en el borrador.
7. **El reporte a stderr es best-effort.** Si el draft se escribió bien y el reporte no, la salida es
   1 con el draft ya entregado; el aviso del fallo no tiene dónde escribirse.
8. **Sin salida en streaming.** El borrador se entrega completo, recién después de verificar.

## 3. Backlog sugerido (no autorizado)

Nada de esto está pedido ni planificado; si algo se prioriza, entra como tarea nueva en el documento
ODD con su spec y sus tests.

- Timeouts explícitos y cancelación con mensaje propio.
- Límite de tamaño de entrada y de cantidad de cambios, con error tipado.
- Manejo de límites de tasa del proveedor (espera y reintento acotado).
- Reporte de uso (tokens/costo) para decisiones de modelo.
- Modo `-check` que valide la petición sin llamar al modelo.

## 4. Hallazgos del review nativo

| ID | Lente | Severidad | Estado |
| --- | --- | --- | --- |
| `R3-001` | reliability | CRITICAL | corregido: los tokens protegidos se comparan byte a byte en el texto y en el borrador |
| `R4-001` | resilience | WARNING (advisory, informacional) | corregido: toda escritura de salida se chequea y una entrega truncada sale 1 |

## 5. Fuente de verdad y cómo verificar

- Specs (`S1`–`S12`), tareas y evidencia commit por commit: `odd/tasks/text-correction-v1.md`.
- Contrato de la CLI, flags, códigos de salida y formato JSON: `README.md`.
- Verificación local: `make check` (gofmt, vet, build, test) — no necesita red ni credencial.
- Prueba manual contra el modelo: ver "Quick start" en el `README.md`.
