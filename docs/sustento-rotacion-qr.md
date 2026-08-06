# Rotación condicional como habilitador de la factorización QR

Sustento de la decisión de diseño · Reto técnico Interseguro, División TI

---

> **Tesis**
>
> El enunciado menciona **rotación** en la arquitectura de la solución y **factorización QR** en la funcionalidad requerida. No son operaciones alternativas ni redundantes: la factorización QR impone una restricción de forma, y la rotación es precisamente el mecanismo que permite satisfacerla. Por eso la rotación se aplica **de forma condicional**, como paso habilitante y no decorativo.

---

## 0 · La restricción que define todo el diseño

La factorización QR solo está definida cuando la matriz tiene al menos tantas filas como columnas:

```
filas ≥ columnas
```

Es una restricción matemática, no una limitación de la librería. Verificado en el código fuente de `gonum/mat`:

```go
// Factorize computes the QR factorization of an m×n matrix a where m >= n.
func (qr *QR) factorize(a Matrix, norm lapack.MatrixNorm) {
	m, n := a.Dims()
	if m < n {
		panic(ErrShape)
	}
	...
}
```

El enunciado, en cambio, admite matrices **rectangulares** sin restricción de forma. Existe entonces un conjunto de entradas válidas según el enunciado para las cuales la operación requerida no está definida.

Toda matriz de entrada cae en uno de dos casos, y el sistema decide el camino a partir de esa única comparación.

---

## Caso A · La matriz ya cumple: se factoriza directo

Entrada de 3×2:

```
3   0
4   5
0   0
```

```
3 filas ≥ 2 columnas  ✓
no se rota · se factoriza tal cual
```

La matriz del usuario llega intacta a la factorización.

---

## Caso B · La matriz es ancha: la rotación la habilita

Entrada de 2×3:

```
0   5   0
3   4   0
```

Tras rotar 90° en sentido horario, de 2×3 pasa a 3×2:

```
3   0
4   5
0   0
```

```
2 filas < 3 columnas  ✗
tras rotar:  3 ≥ 2    ✓
```

Rotar 90° **intercambia las dimensiones**: una matriz de m×n se convierte en una de n×m. Ese intercambio es exactamente lo que convierte una matriz inviable en una viable.

> La rotación no se agrega porque el enunciado la nombra, sino porque resuelve un problema real: sin ella, toda matriz ancha sería irrecibible.

---

## Desde aquí ambos caminos convergen

Los dos casos llegan a la misma matriz de 3×2. El recorrido siguiente usa el método de Gram-Schmidt por su claridad expositiva.

### 1 · Primera columna

Se aísla la primera columna y se trata como una flecha en el espacio:

```
columna 1 = ( 3 , 4 , 0 )
```

Se mide su longitud:

```
longitud = √(3² + 4² + 0²) = √25 = 5
```

Se divide entre su propia longitud para que mida exactamente 1:

```
q₁ = ( 3/5 , 4/5 , 0 ) = ( 0.6 , 0.8 , 0 )
```

La longitud retirada se guarda dentro de R:

```
r₁₁ = 5
```

### 2 · Segunda columna, hecha perpendicular a la primera

```
columna 2 = ( 0 , 5 , 0 )
```

Cuánto de esta columna apunta en la dirección de q₁:

```
r₁₂ = (0)(0.6) + (5)(0.8) + (0)(0) = 4
```

Se le resta esa parte; lo que sobra apunta en una dirección nueva:

```
( 0 , 5 , 0 ) − 4 · ( 0.6 , 0.8 , 0 ) = ( −2.4 , 1.8 , 0 )
```

Se mide y se normaliza igual que antes:

```
longitud = √(2.4² + 1.8²) = √9 = 3
q₂ = ( −0.8 , 0.6 , 0 )
r₂₂ = 3
```

Aquí nacen los decimales y el signo negativo: son consecuencia de dividir entre longitudes y de restar la parte compartida.

### 3 · Tercera dirección para completar Q

Q debe ser cuadrada del tamaño de las filas, es decir 3×3. Las direcciones q₁ y q₂ viven en un mismo plano; falta la perpendicular a ambas:

```
q₃ = ( 0 , 0 , 1 )
```

### 4 · Las dos matrices resultado

```
Q · 3×3                      R · 3×2

 0.6   −0.8    0              5    4
 0.8    0.6    0              0    3
 0      0      1              0    0
```

**Q**: columnas perpendiculares entre sí, cada una de longitud 1.
**R**: triangular superior, todo cero bajo la diagonal.

### 5 · Verificación

```
fila 1 × col 1 = (0.6)(5) + (−0.8)(0) + (0)(0) = 3        ✓
fila 1 × col 2 = (0.6)(4) + (−0.8)(3)          = 0        ✓
fila 2 × col 1 = (0.8)(5) + ( 0.6)(0)          = 4        ✓
fila 2 × col 2 = (0.8)(4) + ( 0.6)(3)          = 5        ✓
```

Se recupera exactamente la matriz de partida.

**El punto crítico del diseño:**

| Caso | Qué reconstruye `Q × R` |
|---|---|
| **A** | La matriz que el usuario envió. La factorización corresponde literalmente a su entrada. |
| **B** | La matriz **rotada**, no la original. Por eso la respuesta devuelve también la rotada y lo declara en `factorizedFrom`. |

> Esta transparencia es deliberada: quien reciba la respuesta puede verificar `Q × R` por su cuenta y obtener exactamente la matriz declarada.

### 6 · Los valores que viajan a la segunda API

```
Q →  0.6   −0.8    0    0.8    0.6    0    0    0    1
R →  5      4      0    3      0      0
```

Quince valores en total.

```
suma     = 2.2 + 12 = 14.2
promedio = 14.2 ÷ 15 = 0.947
```

| Métrica | Valor |
|---|---|
| máximo | 5 |
| mínimo | −0.8 |
| suma | 14.2 |
| promedio | 0.947 |
| ¿diagonal? | false |

`isAnyDiagonal` es falso porque Q tiene un −0.8 fuera de la diagonal y R tiene un 4.

---

## La salida real difiere en los signos — y también es correcta

El recorrido anterior usa Gram-Schmidt. La implementación delega en `gonum`, que emplea reflexiones de Householder: numéricamente más estables, pero con distinta elección de signos.

```
Gram-Schmidt · recorrido manual     Householder · salida de gonum

Q =  0.6   −0.8    0                Q = −0.6   −0.8    0
     0.8    0.6    0                    −0.8    0.6    0
     0      0      1                     0      0      1

R =  5      4                       R = −5     −4
     0      3                            0      3
     0      0                            0      0
```

Se invierte el signo de la primera columna de Q, compensado por el signo de la primera fila de R.

La factorización QR **no es única**: está determinada salvo el signo de cada columna de Q, siempre que la fila correspondiente de R lo compense. Ambas satisfacen la identidad fundamental:

```
(−0.6)(−5) + (−0.8)(0) + (0)(0) = 3                ✓
(−0.6)(−4) + (−0.8)(3) + (0)(0) = 2.4 − 2.4 = 0    ✓
(−0.8)(−5) + ( 0.6)(0) + (0)(0) = 4                ✓
(−0.8)(−4) + ( 0.6)(3) + (0)(0) = 3.2 + 1.8 = 5    ✓
```

Las estadísticas resultantes difieren porque los valores difieren, y ambas son correctas para su respectiva factorización:

| Métrica | Gram-Schmidt | Householder · salida real |
|---|---|---|
| máximo | 5 | 3 |
| mínimo | −0.8 | −5 |
| suma | 14.2 | −6.6 |
| promedio | 0.947 | −0.44 |
| ¿diagonal? | false | false |

> Por esta razón las pruebas verifican la reconstrucción **Q × R = A** y nunca valores literales de Q y R: una prueba escrita contra números fijos fallaría al cambiar de algoritmo, sin que nada estuviera mal.

---

## Alternativas evaluadas y descartadas

| Alternativa | Por qué se descartó |
|---|---|
| Rechazar matrices anchas con HTTP 400 | El enunciado las admite explícitamente al hablar de matrices rectangulares. Rechazarlas reduce el alcance solicitado. |
| Transponer en lugar de rotar | Matemáticamente equivalente en cuanto al cambio de dimensiones y algo más canónico, pero no está mencionada en el enunciado. La rotación sí lo está. |
| Aplicar factorización LQ a matrices anchas | Es la operación hermana correcta para el caso `filas < columnas`, pero introduce una segunda descomposición no solicitada y complica el contrato de salida. |
| Rotar siempre, antes de todo | Rompe el caso que ya funcionaba: una matriz 3×2 rotada se convierte en 2×3, volviéndose inviable. La rotación debe ser condicional. |

---

## Puntos de defensa

1. **La ambigüedad del enunciado se detectó y se resolvió con criterio.** No se descartó ninguna de las dos menciones; se encontró la relación funcional entre ambas.

2. **La restricción de forma se verificó en la fuente,** no se asumió. El código de `gonum/mat` confirma que la factorización exige `filas ≥ columnas`.

3. **La rotación tiene un propósito.** Es el mecanismo que habilita el procesamiento de matrices anchas, no un paso agregado para cumplir con una frase del enunciado.

4. **La transformación se declara explícitamente.** El campo `factorizedFrom` en la respuesta indica sobre qué matriz se calculó Q y R, de modo que la identidad `Q × R` sea verificable por el consumidor.

5. **Se evaluaron alternativas y se documentó por qué se descartaron:** rechazar la entrada, transponer, o usar factorización LQ.

6. **Los signos de Q y R pueden diferir según el algoritmo.** Gonum emplea reflexiones de Householder, numéricamente más estables que Gram-Schmidt; ambas producen factorizaciones válidas que satisfacen `Q × R = A`, aunque con signos distintos. Por eso las pruebas verifican la reconstrucción del producto y no valores literales.

---

*El recorrido numérico usa el método de Gram-Schmidt por su claridad expositiva. La implementación delega en gonum, que emplea reflexiones de Householder: ambas satisfacen `Q × R = A`.*
