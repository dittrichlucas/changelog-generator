## Roadmap

### Scaffold

- [x] Configurar golang
- [x] Adicionar pacote do github api para golang
- [x] Adicionar sdk do github actions para golang
- [x] Adicionar issues no repositório do github

### MVP

- [x] Definir regra de negócio para geração do changelog
    - [x] Issues
    - [x] PRs
- [x] Preencher template da documentação baseado nos valores da nova release
    - [x] Issues
    - [x] PRs
- [ ] Validações
    - [x] Caso o arquivo não exista, deve cria-lo
    - [x] Se não houver dados na release atual, o changelog não deve ser gerado
    - [x] Se não houver responsável pela PR, adicionar o criador no template do changelog
    - [ ] Se for a primeira release, dispensar necessidade da release anterior
- [x] Acrescentar CHANGELOG da nova release no início do arquivo existente
- [x] Transformar o script final em um action

### Improvements

- [ ] Traduzir arquivos para inglês
- [ ] Gerar log do processamento dos passos
- [ ] Gerar changelog retroativo
    - [ ] Todas as tags
    - [ ] De/até tag
- [ ] Refatorações
    - [ ] Retornar erros corretamente
    - [ ] Renomear funções e variáveis corretamente
- [ ] Novas features
    - [ ] Personalizar changelog (P&D)
    - [ ] Adicionar o changelog gerado na descrição da release
- [x] Escolher nome da action
- [ ] Adicionar testes unitários e integrados
- [ ] Adicionar fluxo de CI/CD
    - [ ] Lint
    - [ ] Build
    - [ ] Testes

### Business rule

- [x] Para issue:
    - [x] A data de início deve ser a data de publicação da release anterior
    - [x] A data fim deve ser a data de publicação da próxima release
    - [x] Deve retornar apenas issues fechadas
    - [x] O valor que deve retornar por issue deve ser o título

- [x] Para PR:
    - [x] A data de início deve ser a data de publicação da release anterior
    - [x] A data fim deve ser a data de publicação da próxima release
    - [x] Deve retornar apenas PR mergeado
    - [x] O valor que deve retornar por PR deve ser, inicialmente, o título e, após, o valor do changelog retornado no body
