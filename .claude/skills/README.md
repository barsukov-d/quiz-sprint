# Quiz Sprint Skills

Custom skills для улучшения качества разработки Quiz Sprint TMA проекта.

## Установленные Skills

### 1. golang-pro.md
**Senior Go Developer Expertise**

Специализация:
- Идиоматический Go код (Go 1.21+)
- Concurrency patterns (goroutines, channels, select, sync)
- Performance optimization и memory management
- Testing best practices (table-driven tests, benchmarks)
- Error handling patterns
- Context-aware programming

**Когда использовать:**
```
Оптимизируй quiz repository используя golang-pro patterns
```

### 2. ddd-architecture.md
**Domain-Driven Design & Clean Architecture**

Специализация:
- DDD принципы (aggregates, value objects, domain events)
- Clean Architecture layers (domain, application, infrastructure, presentation)
- Hexagonal Architecture (ports & adapters)
- Repository pattern
- Use Case pattern
- Dependency inversion

**Когда использовать:**
```
Применяя DDD principles, предложи улучшения для quiz domain model
```

### 3. backend-patterns.md
**Backend Development Patterns**

Специализация:
- RESTful API design (HTTP methods, status codes, versioning)
- Authentication/Authorization (JWT, OAuth 2.0, API keys)
- Database patterns (repository, unit of work, query optimization)
- Caching strategies (cache-aside, write-through, write-behind)
- Scalability patterns (horizontal/vertical scaling, microservices)
- Security (OWASP Top 10, input validation, secrets management)
- Monitoring & Observability (metrics, tracing, health checks)

**Когда использовать:**
```
Используя backend-patterns, помоги спроектировать API для leaderboard
```

## Как работают Skills

Skills автоматически активируются когда:
- Вы работаете с Go кодом (golang-pro)
- Проектируете архитектуру или domain model (ddd-architecture)
- Работаете с API, базами данных, кешированием (backend-patterns)

Вы также можете явно запросить skill:
```
Используя golang-pro skill, оптимизируй конкурентный доступ к session repository
```

## Примеры использования

### Оптимизация Go кода
```
Проверь quiz_handler.go на соответствие Go best practices
```
*Автоматически применит golang-pro skill*

### Архитектурные решения
```
Предложи улучшения для DDD структуры backend проекта
```
*Автоматически применит ddd-architecture skill*

### API Design
```
Помоги спроектировать RESTful API для quiz sessions
```
*Автоматически применит backend-patterns skill*

### Комбинированное использование
```
Используя golang-pro и ddd-architecture skills, реализуй 
aggregate для Quiz с правильными concurrency patterns
```

## Дополнительные Skills

Проект также использует skills из плагина `javascript-typescript`:
- **modern-javascript-patterns** - ES6+, async/await, functional patterns
- **typescript-advanced-types** - Generics, conditional types, utility types
- **nodejs-backend-patterns** - Express/Fastify, middleware, auth
- **javascript-testing-patterns** - Jest, Vitest, Testing Library

## Добавление новых Skills

Чтобы добавить новый skill:

1. Создайте файл `.claude/skills/your-skill-name.md`
2. Опишите expertise и patterns
3. Укажите "When to Use This Skill"
4. Добавьте примеры кода
5. Commit и push

Skills автоматически подхватятся Claude Code при следующем запуске.

## Полезные ресурсы

- [Claude Code Skills](https://claude-plugins.dev/skills) - Community skills registry
- [Awesome Claude Code](https://github.com/hesreallyhim/awesome-claude-code) - Curated skills collection
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) - Go best practices
- [Domain-Driven Design Reference](https://www.domainlanguage.com/ddd/reference/) - DDD patterns
