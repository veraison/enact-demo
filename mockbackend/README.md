# Mock Backend
This project is a minimal backend, which should have the same endpoints as the enact-demo backend.
The purpose here is to be able to run this 0 dependency minimal backend while running the agent.

To run the backend, go into `/mockbackend` and run `go run ./main.go`. You will see console output confirming the backend is running at port `8000`.

When making requests to `mockbackend`, the params you sent will be output to the console in the terminal.


## Parsing Request body as JSON 

```
    if err := c.BindJSON(&reqBody); err != nil {
        log.Println(err.Error())
    }

    c.JSON(200, gin.H{
        "name_1": reqBody.AK_PUB,
        "name_2": reqBody.EK_PUB,
    })
```