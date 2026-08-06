# Requerimientos generales del sistema

Reto técnico Interseguro · División TI

Requerimientos transversales a ambas APIs. Los específicos de cada servicio están en `fiber-api/docs/requerimientos-paso-1.md` y `express-api/docs/requerimientos-paso-2.md`.

La columna **fase** indica cuándo se aborda cada requerimiento: *local* durante el desarrollo, *cierre* una vez que ambas APIs funcionan en la máquina.

---

## 1. Alcance

Sistema compuesto por dos APIs REST independientes que se comunican por HTTP. La primera procesa una matriz mediante factorización QR; la segunda calcula estadísticas descriptivas sobre las matrices resultantes.

**Origen:** pág. 2 — "Arquitectura de la solución"

---

## 2. Arquitectura

```
Cliente
   │
   │ POST /api/v1/matrix/process
   ▼
fiber-api (Go · Fiber · puerto 3000)
   │  valida → rota si hace falta → factoriza QR
   │
   │ POST /api/v1/statistics
   ▼
express-api (Node · Express · puerto 4000)
   │  calcula max, min, average, sum, isAnyDiagonal
   │
   └──► respuesta de vuelta a fiber-api, que compone el resultado final
```

El cliente nunca invoca a express-api directamente. fiber-api actúa como cliente HTTP de express-api.

---

## 3. Contrato entre servicios

### fiber-api → express-api

```
POST /api/v1/statistics
Content-Type: application/json

{
  "matrices": [
    { "name": "Q", "data": [[0.6, -0.8, 0], [0.8, 0.6, 0], [0, 0, 1]] },
    { "name": "R", "data": [[5, 4], [0, 3], [0, 0]] }
  ]
}
```

### express-api → fiber-api

```
200 OK

{
  "max": 5,
  "min": -0.8,
  "sum": 14.2,
  "average": 0.947,
  "isAnyDiagonal": false
}
```

Las estadísticas se calculan sobre el conjunto unificado de valores de todas las matrices recibidas, no por matriz individual.

---

## 4. Requerimientos no funcionales transversales

### RNG-01 · Frameworks obligatorios

**Origen:** pág. 2 — "los frameworks Fiber para la API en Go y Express.js para la API en Node.js"
**Fase:** local

- La API en Go usa Fiber v2. Se descarta v3 por encontrarse en beta.
- La API en Node.js usa Express 4.

---

### RNG-02 · Comunicación por HTTP

**Origen:** pág. 2 — "Implementar la comunicación entre las dos API utilizando un mecanismo como HTTP"
**Fase:** local

**Criterios de aceptación**

- fiber-api actúa como cliente HTTP de express-api.
- La invocación tiene un tiempo máximo de espera configurable.
- Si express-api no responde o devuelve error, fiber-api entrega un mensaje descriptivo sin exponer detalles internos.

---

### RNG-03 · Configuración externalizada

**Origen:** derivado de RNG-04 y RNG-05
**Fase:** local

La configuración que varía entre entornos se inyecta por variables de entorno, con valores por defecto aptos para desarrollo local.

**Criterios de aceptación**

- La URL de express-api se lee de `STATISTICS_API_URL`, con `http://localhost:4000` por defecto.
- Los puertos se leen de `PORT`.
- Ningún valor de entorno se versiona en el repositorio.

**Nota de diseño.** Este requerimiento se cumple desde la primera línea de código aunque la contenerización venga después. Dentro de Docker, express-api se resuelve por nombre de servicio y no por `localhost`; leer la URL de una variable evita tener que modificar código al dockerizar.

---

### RNG-04 · Contenerización

**Origen:** pág. 2 — "Utilizar Docker para contenerizar las aplicaciones y facilitar su despliegue en diferentes entornos"
**Fase:** cierre

**Criterios de aceptación**

- Cada servicio cuenta con su propio `Dockerfile`.
- Las imágenes se construyen en múltiples etapas para minimizar su tamaño.
- Los contenedores se ejecutan con un usuario sin privilegios.
- Existe un `docker-compose.yml` en la raíz que levanta ambos servicios en una red común.
- La comunicación entre servicios se resuelve por nombre de servicio, no por IP.
- Cada servicio expone `GET /health`, consumido por el healthcheck del contenedor.

---

### RNG-05 · Despliegue en la nube

**Origen:** pág. 2 — "Utilizar servicios en la nube para la implementación y el despliegue de las aplicaciones"
**Fase:** cierre

**Criterios de aceptación**

- Ambos servicios accesibles públicamente mediante URL.
- La configuración sensible se inyecta por variables de entorno del proveedor.

---

### RNG-06 · Documentación

**Origen:** pág. 2 — "Documentar el código de manera clara y concisa, siguiendo las mejores prácticas de codificación"
**Fase:** local y cierre

**Criterios de aceptación**

- Todo identificador exportado documentado según la convención de su lenguaje.
- README con instalación, ejecución y ejemplos de petición y respuesta.
- Decisiones de interpretación del enunciado documentadas con su justificación.

---

### RNG-07 · Coherencia estructural

**Origen:** pág. 4 — "No hay un estándar específico para los nombres de los objetos creados, pero se espera coherencia en su estructura y documentación"
**Fase:** local

**Criterios de aceptación**

- Convenciones de nombres uniformes en ambos servicios.
- Estructura de carpetas equivalente, adaptada a las convenciones de cada lenguaje.
- Formato de respuesta y de error homogéneo entre ambas APIs.
- Versionado de rutas bajo `/api/v1` en ambos servicios; `/health` sin versionar.
- Campos JSON en `camelCase`.

---

### RNG-08 · Corrección y eficiencia

**Origen:** pág. 3 — "de manera eficiente y correcta" · pág. 4 — "Se valorará la eficiencia y la elegancia de la solución"
**Fase:** local

**Criterios de aceptación**

- La factorización delega en una librería numérica establecida en lugar de una implementación propia.
- Las transformaciones de matrices operan en una sola pasada sobre los elementos.
- Sin dependencias innecesarias.
- Los valores numéricos se tratan como flotantes de doble precisión en ambos servicios: la factorización QR produce decimales y negativos aunque la entrada sea de enteros positivos.

---

## 5. Requerimientos opcionales

**Origen:** pág. 3 — "Funcionalidad opcional"
**Fase:** cierre

### RGO-01 · Frontend

Interfaz que consuma ambas APIs y presente los resultados de la factorización y las estadísticas.

### RGO-02 · Autenticación JWT

- Los endpoints de procesamiento requieren un token válido.
- `/health` permanece público.
- El algoritmo de firma se valida contra una lista explícita; nunca se acepta `none`.
- Los tokens tienen expiración.

### RGO-03 · Pruebas

- Pruebas unitarias sobre la lógica de negocio de ambos servicios.
- Prueba de integración que verifique el flujo completo entre ambas APIs.

---

## 6. Ambigüedades del enunciado y su resolución

El enunciado indica en la pág. 4 que ante dudas se espera que el candidato tome decisiones informadas y las sustente. Esta sección constituye ese sustento.

| Ambigüedad | Resolución adoptada |
|---|---|
| La pág. 2 describe la operación de la API en Go como rotación; la pág. 3 como factorización QR. | Se implementan ambas: la rotación como paso habilitante condicional de la factorización, dado que QR exige `filas ≥ columnas` y la rotación intercambia dimensiones. |
| El enunciado admite matrices rectangulares, pero la factorización QR no está definida para matrices con más columnas que filas. | La rotación resuelve el conflicto sin rechazar entradas que el enunciado declara válidas. |
| No se especifica el ángulo ni el sentido de la rotación. | 90° en sentido horario, por ser la interpretación convencional y la que produce el intercambio de dimensiones requerido. |
| No se especifica si las estadísticas se calculan por matriz o sobre el conjunto. | Sobre el conjunto unificado, siguiendo la redacción en plural del enunciado. |
| No se especifica si "verificar si alguna matriz es diagonal" devuelve un valor o varios. | Un único booleano para el conjunto, siguiendo la redacción. |
| No se especifica el destinatario de la respuesta de la API de estadísticas. | Responde a fiber-api, que compone la respuesta final para el cliente, según el flujo descrito en la arquitectura. |

Sustento detallado de la primera y más relevante en `sustento-rotacion-qr.html`.

---

## 7. Trazabilidad

| Requerimiento | Origen | Fase | Prioridad |
|---|---|---|---|
| RNG-01 Frameworks | pág. 2 | local | Obligatorio |
| RNG-02 Comunicación HTTP | pág. 2 | local | Obligatorio |
| RNG-03 Configuración externalizada | derivado | local | Obligatorio |
| RNG-04 Contenerización | pág. 2 | cierre | Obligatorio |
| RNG-05 Despliegue en la nube | pág. 2 | cierre | Obligatorio |
| RNG-06 Documentación | pág. 2 | ambas | Obligatorio |
| RNG-07 Coherencia estructural | pág. 4 | local | Obligatorio |
| RNG-08 Corrección y eficiencia | pág. 3, 4 | local | Obligatorio |
| RGO-01 Frontend | pág. 3 | cierre | Opcional |
| RGO-02 JWT | pág. 3 | cierre | Opcional |
| RGO-03 Pruebas | pág. 3 | cierre | Opcional |
