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
query Test1($id: ID!) {
  todos {
    id
    text
    user {
      id
      name
    }
  }
  todo(id: $id) {
    id
    text
    done
  }
  todos {
    id
    text
    user {
      id
      name
    }
  }
}
```
#### Variables
```json
{
  "id": "T2"
}
```
#### Logs
```json
{
   "GraphQLOperation":"Test1",
   "Payload":{
      "RawQuery":"query Test1($id: ID!) {\n  todos {\n    id\n    text\n    user {\n      id\n      name\n    }\n  }\n  todo(id: $id) {\n    id\n    text\n    done\n  }\n  todos {\n    id\n    text\n    user {\n      id\n      name\n    }\n  }\n}",
      "Variables":{
         "id":"T2"
      }
   },
   "Meta":{
      "Resolvers":{
         "Query:todo":1,
         "Query:todos":1
      },
      "Tags":{
         "hasRole":{
            "Query:todo":1,
            "Todo:done":1,
            "User:name":3
         },
         "lang":{
            
         }
      }
   }
}
```