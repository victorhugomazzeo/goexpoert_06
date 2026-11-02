# Clima por CEP com Observabilidade

Dois serviços em Go que recebem um CEP e retornam a cidade e a temperatura atual em Celsius, Fahrenheit e Kelvin, com tracing distribuído via OpenTelemetry e Zipkin.

## Arquitetura

```
curl ──POST──▶ service-a (:8080) ──GET──▶ service-b (:8081) ──▶ ViaCEP
               valida o CEP               busca cidade e      └─▶ WeatherAPI
                                          temperatura

service-a, service-b ──OTLP──▶ otel-collector (:4317) ──▶ zipkin (:9411)
```

- **service-a**: recebe o POST, valida o CEP e encaminha para o service-b.
- **service-b**: consulta a cidade no ViaCEP, a temperatura na WeatherAPI e faz as conversões.

## Pré-requisitos

- Docker com Docker Compose
- Uma chave gratuita da [WeatherAPI](https://www.weatherapi.com/)

## Como rodar

1. Crie um arquivo `.env` na raiz do projeto:

   ```
   WEATHER_API_KEY=sua_chave
   ```

2. Suba tudo:

   ```bash
   docker compose up --build
   ```

## Como fazer a requisição

```bash
curl -i -X POST http://localhost:8080/weather \
  -H "Content-Type: application/json" \
  -d '{"cep":"01001000"}'
```

No PowerShell, use `curl.exe` e passe tudo numa linha:

```powershell
curl.exe -i -X POST http://localhost:8080/weather -d '{"cep":"01001000"}'
```

| Cenário | Corpo enviado | Status | Resposta |
| --- | --- | --- | --- |
| Sucesso | `{"cep":"01001000"}` | 200 | `{"city":"São Paulo","temp_C":28.5,"temp_F":83.3,"temp_K":301.65}` |
| CEP com formato inválido | `{"cep":"123"}` | 422 | `invalid zipcode` |
| CEP que não é string | `{"cep":29902555}` | 422 | `invalid zipcode` |
| CEP não encontrado | `{"cep":"99999999"}` | 404 | `can not find zipcode` |

## Como ver os traces no Zipkin

1. Faça ao menos uma requisição.
2. Abra http://localhost:9411.
3. Clique em **Run Query** e depois em **Show** no trace desejado.

Cada requisição gera um único trace com os spans:

```
service-a: post /weather
└─ service-a: http get                 (chamada ao service-b)
   └─ service-b: get /weather
      ├─ service-b: busca-cep          (span manual: ViaCEP)
      │  └─ service-b: http get
      └─ service-b: busca-clima        (span manual: WeatherAPI)
         └─ service-b: http get
```

Os spans manuais trazem os atributos `cep`, `city` e `temp_c`, e ficam marcados como erro quando a API externa falha.

## Decisões

- **Kelvin:** o enunciado indica `K = C + 273`, mas o exemplo de resposta (28.5 °C → 301.65 K) usa `273.15`. Foi adotado `273.15`.
- **service-b indisponível:** o service-a responde `502 Bad Gateway`.

## Testes

```bash
go test ./...
```
