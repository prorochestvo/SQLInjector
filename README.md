# SQLInjector
SQLInjector is tiny project designed to simplify SQL database operations. 
It build upon on [sqlboiler](https://github.com/volatiletech/sqlboiler) orm and enhance workflow around it for supports database migrations and integrates the OData filtering.

## Features
- [ ] migration
- [ ] sqlbuilder
- [ ] transaction
- [ ] odata filter
- [ ] logging


## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites
- Go (1.22+)
- A supported SQL database (PostgreSQL, MySQL, SQLite, etc.)
- `sqlboiler` CLI tool
- `sql-migrate` for migrations
- `odatafilter` package for filtering

### Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/yourusername/SQLInjector.git
    cd SQLInjector
    ```

2. Install dependencies:
    ```sh
    go get -u github.com/volatiletech/sqlboiler/v4
    go get -u github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql
    go get -u github.com/rubenv/sql-migrate/...
    go get -u github.com/yourusername/odatafilter
    ```

3. Set up your `sqlboiler` configuration:
   Create a `sqlboiler.toml` file in the root directory of your project with your database configuration.

    ```toml
    [psql]
    dbname = "your_db"
    host = "localhost"
    user = "your_user"
    pass = "your_password"
    port = 5432
    sslmode = "disable"
    ```

4. Generate models:
    ```sh
    sqlboiler psql
    ```

5. Apply migrations:
   Create a `dbconfig.yml` file for migrations:

    ```yaml
    development:
      dialect: postgres
      datasource: user=your_user dbname=your_db sslmode=disable
    ```

   Apply migrations:
    ```sh
    sql-migrate up
    ```

## Usage

### Running the Project

To start using SQLInjector, build and run the project:

```sh
go build -o SQLInjector
./SQLInjector