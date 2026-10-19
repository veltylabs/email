# Cuánto cuesta enviar correos con Cloudflare Email Service

> **Consultado el 2026-10-19** en la documentación de Cloudflare. Los precios y límites cambian: antes de decidir, confirma en
> <https://developers.cloudflare.com/email-service/platform/pricing/> y <https://developers.cloudflare.com/email-service/platform/limits/>.
> El **envío de correos (Email Sending) está en beta pública** desde abril de 2026; la vía SMTP, en beta desde junio de 2026.

## Lo esencial

| Quién recibe el correo | Plan de Cloudflare necesario | Costo |
|---|---|---|
| **Direcciones verificadas de tu propia cuenta** («destination addresses»): el personal que recibe los avisos | **Cualquiera, incluso Workers Gratis** | **Gratis**, no cuenta para la cuota mensual ni para los límites diarios |
| **Cualquier otra persona** (un paciente, un proveedor) | **Workers de pago** (*Workers Paid*) | **3.000 correos al mes incluidos**; después **US$ 0,35 por cada 1.000** |

- En **Workers Gratis** el envío a destinatarios arbitrarios **no está disponible**.
- Los correos **rechazados** antes de enviarse (por ejemplo, a una dirección en la lista de supresión) **no cuentan** en la cuota.
- La cuota de 3.000 es **por cuenta y por mes**, según tu ciclo de facturación de Cloudflare.
- El precio del plan **Workers de pago** en sí no lo trae esta página: míralo en <https://developers.cloudflare.com/workers/platform/pricing/>.

## Qué significa para la clínica

Los avisos del sistema (stock bajo, producto por vencer, deuda por pagar) van **al personal**, que son unas pocas direcciones fijas. Hay dos caminos:

1. **Costo cero:** agregar cada correo del personal como **dirección de destino verificada** en Cloudflare (*Compute → Email Service → Email Routing → Destination Addresses*). Cada persona recibe un correo y debe pulsar «verificar». Cabe un máximo de **200** direcciones por cuenta. Funciona con la cuenta gratuita.
2. **Con plan de pago:** necesario si algún día se escribe a pacientes o a proveedores. Para el volumen de avisos internos, 3.000 correos al mes incluidos sobran con holgura.

## Lo que hay que preparar (una sola vez)

1. El dominio debe usar **DNS de Cloudflare** (`monjitaschillan.cl` ya lo usa).
2. **Incorporar el dominio** en *Compute → Email Service → Email Sending → Onboard Domain*: Cloudflare añade los registros SPF, DKIM y DMARC en el subdominio `cf-bounce`. Tarda de 5 a 15 minutos, a veces hasta 24 horas.
3. Una cuenta nueva parte con un **límite diario de envío conservador** que sube según el comportamiento y la reputación; se puede pedir un aumento.
4. Un **token de API** con el permiso *Email Sending: Edit*. **Ese token es un secreto:** se entrega por el asistente de despliegue y nunca se guarda en la base de datos, en el repositorio ni en el chat.

## Una cuenta gratuita con tarjeta agregada **no** es el plan de pago

Agregar una tarjeta para usar la base de datos (D1) permite cobrar el uso que excede lo gratuito de **ese** producto; **no** activa el plan *Workers de pago*. Revisa en *Billing* qué plan tiene la cuenta antes de contar con el envío a destinatarios externos.
