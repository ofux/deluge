deluge("Some name", "10s", {
    "sc1": {
        "concurrent": 100,
        "delay": "2s"
    }
});

scenario("sc1", "Some scenario", function () {

    http("Some request", {
        "url": "http://localhost:8080/hello/toto"
    });

});