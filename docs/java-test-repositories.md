# Java Test Repositories

Use this repository set to test Cosha's Java indexing, symbol extraction, search, benchmarking, and Java/Spring-aware behavior across progressively larger codebases.

Do not commit local checkout paths. Configure paths in a gitignored file, such as `.env`:

```sh
COSHA_REPO_SPRING_PETCLINIC=/path/to/spring-petclinic
COSHA_REPO_SPRING_BOOT=/path/to/spring-boot
COSHA_REPO_SPRING_FRAMEWORK=/path/to/spring-framework
COSHA_REPO_KAFKA=/path/to/kafka
COSHA_REPO_HADOOP=/path/to/hadoop
COSHA_REPO_DUBBO=/path/to/dubbo
COSHA_REPO_SHARDINGSPHERE=/path/to/shardingsphere
COSHA_REPO_ELASTICSEARCH=/path/to/elasticsearch
```

## Progression

Recommended order:

1. Spring PetClinic
2. Spring Boot
3. Spring Framework
4. Apache Kafka
5. Apache Hadoop
6. Apache Dubbo
7. Apache ShardingSphere
8. Elasticsearch

Rationale:

- Start with a small Spring app for correctness.
- Move to Spring Boot and Spring Framework for realistic Java/Spring behavior.
- Use Kafka for mixed JVM and Gradle complexity.
- Use Hadoop, ShardingSphere, and Elasticsearch for large-repo stress.
- Use Dubbo to validate service/interface-heavy framework code.

## Common Test Flow

Run this flow for every repository:

```sh
cd "$REPO"
cosha detect
cosha index --rebuild
cosha stats
cosha stats --json
cosha search <query>
cosha search <query> --json
cosha symbols <query>
cosha symbols <query> --json
```

Also check:

- No crashes during indexing.
- Java files are discovered correctly.
- Unsupported files are skipped safely.
- Common ignored directories are not indexed.
- Symbols contain useful file paths and line numbers.
- Package names are extracted correctly.
- JSON output is valid.
- Search results are useful to a coding agent.
- Re-indexing does not duplicate or corrupt index data.

## Benchmarking

Use `hyperfine` for repeatable CLI benchmarks:

```sh
brew install hyperfine
```

Benchmark each command family separately:

```sh
cd "$REPO"

hyperfine --warmup 1 --runs 3 \
  'cosha index --rebuild'

hyperfine --warmup 3 --runs 10 \
  'cosha detect >/dev/null' \
  'cosha stats >/dev/null' \
  'cosha stats --json >/dev/null'

hyperfine --warmup 3 --runs 10 \
  'cosha search Controller >/dev/null' \
  'cosha search Controller --json >/dev/null' \
  'cosha symbols Controller >/dev/null' \
  'cosha symbols Controller --json >/dev/null'
```

For large repositories, reduce runs if the index command is expensive:

```sh
hyperfine --warmup 1 --runs 1 'cosha index --rebuild'
```

Report these fields for each repository:

- Repository ID
- Commit SHA tested
- Machine OS/CPU/RAM
- Go version
- Cosha commit SHA
- Java file count
- Symbol count
- Index size
- `detect` mean time
- `index --rebuild` time
- `stats` mean time
- `stats --json` mean time
- Representative `search` mean time
- Representative `symbols` mean time
- Notes on noise, crashes, invalid JSON, or surprising results

## Report Template

```md
## <repository-id>

- Repository commit:
- Cosha commit:
- Machine:
- Go:
- Java files:
- Symbols:
- Index size:

### Benchmark

| Command | Mean | Min | Max | Notes |
| --- | ---: | ---: | ---: | --- |
| cosha detect | | | | |
| cosha index --rebuild | | | | |
| cosha stats | | | | |
| cosha stats --json | | | | |
| cosha search <query> | | | | |
| cosha symbols <query> | | | | |

### Functional Results

- Detection:
- Indexing:
- Search quality:
- JSON validity:
- Ignored/generated directory behavior:
- Practical task outcome:
```

## Spring PetClinic

Repository variable: `COSHA_REPO_SPRING_PETCLINIC`

Why this repo:

- Small, clean, well-known Spring Boot application.
- Good for quickly verifying basic Java and Spring extraction without large-repo noise.

What to test:

- Java file discovery.
- Package extraction.
- Import extraction.
- Class extraction.
- Method extraction.
- Basic annotation extraction.
- Spring MVC controller detection.
- Service/repository style class discovery.
- Test file discovery.
- Search by class name.
- Search by common Spring terms.

Example queries:

- `OwnerController`
- `VisitController`
- `PetController`
- `Controller`
- `Repository`
- `Service`
- `Vet`

Practical task cases:

- Find the controller method that handles creating or updating an owner.
- Locate the repository or service path used to load owners.
- Find the model class for pets and inspect its relationships.

Success criteria:

- Cosha indexes the repo without errors.
- Java files under `src/main/java` and `src/test/java` are discovered.
- Controller, service, repository, and model classes appear in symbol search.
- Searching for known classes like `OwnerController` returns the expected files.
- JSON output is valid and compact.
- Ignored/generated directories are not accidentally indexed.

## Spring Boot

Repository variable: `COSHA_REPO_SPRING_BOOT`

Why this repo:

- Large, real-world Spring ecosystem repository.
- Useful for validating Spring-heavy Java indexing, annotations, configuration classes, tests, and multi-module Gradle structure.

What to test:

- Large Java repository indexing.
- Multi-module Gradle repository traversal.
- Deep package extraction.
- Class/interface/enum/record extraction.
- Annotation extraction.
- Auto-configuration class discovery.
- ConfigurationProperties-related search.
- Test class discovery.
- Search by Spring-specific symbols.
- Search across many nested modules.
- Handling of build/generated/Gradle directories.

Example queries:

- `AutoConfiguration`
- `ConfigurationProperties`
- `SpringApplication`
- `ApplicationContext`
- `RestController`
- `ConditionalOnClass`
- `EnableAutoConfiguration`
- `Binder`

Practical task cases:

- Find where `SpringApplication` is defined and identify related tests.
- Locate auto-configuration classes for a common Spring Boot feature.
- Search for `ConfigurationProperties` and inspect whether annotations and classes are easy to navigate.

Success criteria:

- Cosha indexes the repo without crashing.
- Generated/build directories are skipped.
- Large numbers of Java symbols are extracted.
- Search for `AutoConfiguration` and `ConfigurationProperties` returns relevant Spring Boot files.
- Symbol search works across multiple modules.
- Output remains usable and not overly noisy.
- Re-running index does not create duplicate rows or corrupted index data.

## Spring Framework

Repository variable: `COSHA_REPO_SPRING_FRAMEWORK`

Why this repo:

- Mature enterprise Java framework with deep package structures, many interfaces, abstract classes, annotations, and tests.
- Good for validating framework-style Java code.

What to test:

- Interface extraction.
- Abstract class extraction.
- Method extraction.
- Nested package discovery.
- Annotation extraction.
- Large test tree indexing.
- Search by foundational Spring symbols.
- Handling of mature enterprise Java patterns.
- Discovery of classes spread across many modules.

Example queries:

- `BeanFactory`
- `ApplicationContext`
- `Environment`
- `Resource`
- `TransactionManager`
- `Controller`
- `Component`
- `Nullable`

Practical task cases:

- Find the main `ApplicationContext` interface and nearby implementations.
- Locate `BeanFactory` and inspect related interfaces.
- Search for `Nullable` and verify annotation-heavy files remain readable in results.

Success criteria:

- Cosha discovers Java files across framework modules.
- Interfaces like `BeanFactory` and `ApplicationContext` are searchable.
- Symbol extraction works on classes, interfaces, enums, annotations, and records where present.
- Package names are correctly stored.
- File paths in search results clearly point to the correct module.
- Large test directories do not break indexing.
- JSON output remains structurally consistent.

## Apache Kafka

Repository variable: `COSHA_REPO_KAFKA`

Why this repo:

- Large JVM repository with Java and Scala, Gradle modules, clients, server code, tests, and complex package structure.
- Useful for testing mixed-language tolerance and Java-only indexing behavior.

What to test:

- Java file indexing in a mixed JVM repo.
- Ignoring unsupported languages gracefully.
- Gradle multi-module layout.
- Client/server package discovery.
- Class/interface extraction.
- Method extraction.
- Test class extraction.
- Search by Kafka client/server terms.
- Handling of large test folders.
- Handling of generated/build directories.

Example queries:

- `KafkaConsumer`
- `KafkaProducer`
- `ConsumerConfig`
- `ProducerConfig`
- `AdminClient`
- `TopicPartition`
- `KafkaServer`
- `Streams`

Practical task cases:

- Find the public consumer API and its config class.
- Locate producer-side classes and related tests.
- Search for `TopicPartition` and confirm model/value classes are easy to find.

Success criteria:

- Cosha indexes Java files without failing on Scala or other unsupported files.
- Unsupported files are skipped cleanly.
- `KafkaConsumer`, `KafkaProducer`, and related classes appear in results.
- Test files and production files are both represented where applicable.
- Build/generated folders are not indexed.
- Search results correctly identify Java files and symbols.
- JSON output is valid even for large result sets.

## Apache Hadoop

Repository variable: `COSHA_REPO_HADOOP`

Why this repo:

- Very large Java-heavy repository.
- Good for testing Cosha's ability to handle scale, deep module structures, older and newer Java styles, and large file counts.

What to test:

- Very large Java repo indexing.
- Deep module traversal.
- Large number of files.
- Large number of symbols.
- Package extraction across many modules.
- Class/interface extraction.
- Search by common Hadoop classes.
- Handling of old-style and new-style Java code.
- Ignoring build/generated/vendor-like directories.

Example queries:

- `FileSystem`
- `Configuration`
- `Path`
- `YarnConfiguration`
- `Namenode`
- `Datanode`
- `MapReduce`
- `ResourceManager`

Practical task cases:

- Find the core `FileSystem` abstraction and common implementations.
- Locate `YarnConfiguration` and related resource manager classes.
- Search for MapReduce entry points and verify module paths are clear.

Success criteria:

- Cosha completes indexing without crashing.
- Java files across Hadoop submodules are discovered.
- Common Hadoop symbols like `FileSystem`, `Configuration`, and `Path` are searchable.
- Search results include module-specific file paths.
- The index remains queryable after indexing a very large repo.
- Unsupported/generated/build files are skipped.
- Output remains compact enough for agent consumption.

## Apache Dubbo

Repository variable: `COSHA_REPO_DUBBO`

Why this repo:

- Large Java RPC framework with interfaces, annotations, SPI patterns, tests, and framework-style abstractions.
- Useful for validating service/interface-heavy Java indexing.

What to test:

- Interface-heavy Java extraction.
- Annotation extraction.
- Service/provider/consumer style class discovery.
- SPI-related class discovery.
- Test file indexing.
- Package extraction across modules.
- Search by RPC framework terms.
- Method extraction in service interfaces and implementations.

Example queries:

- `DubboBootstrap`
- `ReferenceConfig`
- `ServiceConfig`
- `ExtensionLoader`
- `Protocol`
- `Registry`
- `Invoker`
- `Filter`

Practical task cases:

- Find the bootstrap entry point and related config classes.
- Locate `ExtensionLoader` and identify SPI-heavy areas.
- Search for `Protocol` or `Invoker` and confirm interface-heavy results are usable.

Success criteria:

- Cosha indexes the repo without errors.
- Interface and implementation classes are extracted.
- Symbols like `DubboBootstrap`, `ServiceConfig`, and `ExtensionLoader` are searchable.
- Annotation-heavy files are handled cleanly.
- Multi-module structure is represented in file paths.
- Search output helps locate service/provider/extension-related code.
- JSON output remains valid and predictable.

## Apache ShardingSphere

Repository variable: `COSHA_REPO_SHARDINGSPHERE`

Why this repo:

- Large Java middleware repository with many modules, parser-related code, distributed components, tests, and complex naming patterns.
- Useful for stressing module traversal and symbol volume.

What to test:

- Very large multi-module Java indexing.
- Deep package and module structure.
- Parser-related Java classes.
- Large symbol volume.
- Test indexing.
- Class/interface/enum extraction.
- Search by distributed database/middleware terms.
- Handling of generated/build directories.

Example queries:

- `ShardingSphere`
- `SQLParser`
- `RouteContext`
- `Database`
- `Rule`
- `Encrypt`
- `Governance`
- `DistSQL`

Practical task cases:

- Find route-related classes starting from `RouteContext`.
- Search parser-related classes and verify deeply nested module paths stay readable.
- Locate rule or encrypt-related abstractions and related tests.

Success criteria:

- Cosha indexes Java files across many modules.
- Search for `ShardingSphere`, `SQLParser`, and `RouteContext` returns useful results.
- Symbol extraction works despite large package depth.
- Generated/build directories are skipped.
- The index remains usable for search after processing a large symbol set.
- Output remains structured and agent-friendly.

## Elasticsearch

Repository variable: `COSHA_REPO_ELASTICSEARCH`

Why this repo:

- Very large modern Java codebase with Gradle, modules, tests, plugins, and complex architecture.
- Useful for stress-testing modern Java project structures.

What to test:

- Very large modern Java repository indexing.
- Gradle/module-heavy layout.
- Plugin directory traversal.
- Class/interface/record/enum extraction.
- Method extraction.
- Test file discovery.
- Search by Elasticsearch domain terms.
- Handling of generated/build folders.
- Handling of large nested package structures.

Example queries:

- `Elasticsearch`
- `ClusterService`
- `TransportService`
- `IndexService`
- `SearchService`
- `QueryBuilder`
- `RestHandler`
- `Plugin`

Practical task cases:

- Find cluster service classes and related tests.
- Search for REST handler implementations and confirm plugin paths are visible.
- Locate query builder abstractions and check whether records/enums/interfaces are captured.

Success criteria:

- Cosha indexes the repo without crashing.
- Java files in core, modules, plugins, and tests are discovered where applicable.
- Known symbols like `ClusterService` and `TransportService` are searchable.
- Search results identify meaningful module/plugin paths.
- Unsupported/generated/build outputs are skipped.
- JSON output remains valid and not excessively verbose.
- Re-indexing does not corrupt the local index.
