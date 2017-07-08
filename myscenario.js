deluge("Some name", {
    "product": {
        "concurrent": 100,
        "delay": "2s",
        "args": {
            "baseUrl": "http://127.0.0.1:8080"
        }
    }
});

scenario("product", "Test the product entity", function (args, session) {

    let baseUrl = "http://127.0.0.1:8080";
    if (args["baseUrl"]) {
        baseUrl = args["baseUrl"];
    }

    let httpCommonHeaders = {
        "Accept": "application/json",
        "Accept-Encoding": "gzip, deflate",
        "Accept-Language": "fr,fr-fr;q=0.8,en-us",
        "Connection": "keep-alive",
        "User-Agent": "Mozilla/5.0 etc."
    };


    let res1 = http("First unauthenticated request", {
        "url": baseUrl + "/api/v1/account",
        "headers": httpCommonHeaders
    });

    assert(res1["status"] == 401);

    pause("10s");

    let access_token = session["access_token"];
    if (access_token == null) {
        access_token = authenticate();
        session["access_token"] = access_token;
    }

    pause("1s");

    let res2 = http("Authenticated request", {
        "url": baseUrl + "/api/v1/account",
        "headers": merge(httpCommonHeaders, {
            "Authorization": "Bearer " + access_token
        })
    });

    assert(res2["status"] == 200);

    pause("10s");

    for (let i=0; i < 2; i++) {
        let res3 = http("Get all products", {
            "url": baseUrl + "/api/v1/products",
            "headers": merge(httpCommonHeaders, {
                "Authorization": "Bearer " + access_token
            })
        });

        assert(res3["status"] == 200);

        let res4 = http("Create new product", {
            "url": baseUrl + "/api/v1/products",
            "method": "POST",
            "headers": merge(httpCommonHeaders, {
                "Authorization": "Bearer " + access_token
            }),
            "body": toJson({
                "ref": "SJ5",
                // etc
            })
        });

        assert(res4["status"] == 201);
    }

});




let authenticate = function () {
    let resAuth = http("Authentication", {
        "url": "https://sgconnect.com/oauth/token",
        "headers": httpCommonHeaders,
        "form-params": {
            "username": "admin",
            "password": "admin",
            "grant_type": "password"
            // etc...
        }
    });

    assert(resAuth["status"] == 200);
    let access_token = parseJson(resAuth["body"])["access_token"];
    assert(len(access_token) > 0);

    return access_token;
};