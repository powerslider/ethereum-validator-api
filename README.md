# ethereum-validator-api

## Development Setup

**Step 0.** Install [pre-commit](https://pre-commit.com/):

```shell
pip install pre-commit

# For macOS users.
brew install pre-commit
```

Then run `pre-commit install` to setup git hook scripts.
Used hooks can be found [here](.pre-commit-config.yaml).

______________________________________________________________________

NOTE

> `pre-commit` aids in running checks (end of file fixing,
> markdown linting, go linting, runs go tests, json validation, etc.)
> before you perform your git commits.

______________________________________________________________________

**Step 1.** Install external tooling (golangci-lint, etc.):

```shell script
make install
```

**Step 2.** Setup project for local testing (init env file, code lint, test run, etc.):

```shell script
make all
```

**Step 3.** Edit newly created `.env` file with the needed values:

```shell
vim <project_root>/server/.env
```

**Step 4.** Run server:

```shell
make run-server
```

and open http://0.0.0.0:8080/swagger/index.html#/ to open Swagger UI.

______________________________________________________________________

NOTE

> Check [Makefile](Makefile) for other useful commands.

______________________________________________________________________

### Docker Setup

Run server with:

```shell
make compose-up
```

and open http://0.0.0.0:8080/swagger/index.html#/ to open Swagger UI.

Also run:

```shell
make compose-down
```

to clean up the environment.
