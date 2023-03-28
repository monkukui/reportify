# gqlgen-usage-analysis
This is a POC of gathering data of schema usage in gqlgen framework

## Run the server
### Install dependencies
```
$ go mod download
```

### Run server
```
$ go run server.go
```

### Test with GraphQL queries
go to http://localhost:8080/ for GraphQL playground.

### Usage logs
#### Query
```graphql
query Test1($id: ID!, $id2: ID!) {
    todos {
        id
        text
        user {
            id
            name
        }
    }
    todoT1: todo(id: $id) {
        id
        text
        done
    }
    todoT2: todo(id: $id2) {
        id
        text
        done
    }
}
```
#### Variables
```json
{
  "id": "T1",
  "id2": "T2"
}
```
#### Logs
```json
{
  "GraphQLOperation":"Test1",
  "Payload":{
    "RawQuery":"query Test1($id: ID!, $id2: ID!) {\n  todos {\n    id\n    text\n    user {\n      id\n      name\n    }\n  }\n  todoT1: todo(id: $id) {\n    id\n    text\n    done\n  }\n  todoT2: todo(id: $id2) {\n    id\n    text\n    done\n  }\n}",
    "Variables":{
      "id":"T1",
      "id2":"T2"
    }
  },
  "Meta":{
    "Resolvers":{
      "Query:todo":2,
      "Query:todos":1
    },
    "Tags":{
      "hasRole":{
        "Query:todo":2,
        "Todo:done":2,
        "User:name":3
      },
      "lang":{

      }
    }
  }
}
```