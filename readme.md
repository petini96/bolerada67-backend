# Projeto Bolerada67 (Backend)

Este projeto consiste em uma API RESTful desenvolvida para atender empresas de pequeno porte, oferecendo funcionalidades semelhantes a um ERP. Ele engloba gerenciamento de estoque, produtos, fornecedores, controle de clientes, fluxo de caixa, encomendas, entre outras funcionalidades. Além disso, inclui a capacidade de disponibilizar imagens de terceiros, que podem ser obtidas a partir do Google Photos ou Google Drive.

## Tecnologias Utilizadas

- **Backend**: Go (Golang) com o framework Gin, GORM e Viper.
- **Infraestrutura**: Nginx com proxy reverso, Docker, Docker Compose, e SSL (HTTPS).
- **Banco de Dados**: Redis, PostgreSQL e planilhas (Spreadsheet).

## Funcionalidades

### API REST - Google Drive

Esta funcionalidade permite varrer imagens de pastas públicas no Google Drive com o objetivo de realizar download e armazenamento em cache.

### Scrapping Fish

Utiliza a técnica de scraping para extrair dados de uma página chinesa de camisas de time (Yupoo). O objetivo é obter apenas os links de determinadas imagens com base no `folder_id` obtido na URL do site.

### API REST - Google Sheets

Utilizada para importar dados de uma planilha no Google Sheets e, em seguida, incluí-los no banco de dados.

## Configuração do Ambiente

### Backend

- Linguagem: Go (Golang)
- Framework: Gin
- ORM: GORM
- Configuração: Viper

### Infraestrutura

- Servidor Web: Nginx com proxy reverso
- Contêinerização: Docker
- Orquestração: Docker Compose
- Segurança: SSL (HTTPS)

### Banco de Dados

- Cache: Redis
- Banco de Dados Relacional: PostgreSQL
- Armazenamento de Dados: Planilhas (Spreadsheet)

