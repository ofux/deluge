deluge("Some name", {
    "sc1": {
        "concurrent": 100,
        "delay": "2s"
    }
});

scenario("sc1", "Some scenario", function () {

    http({
        "url": "http://localhost:8080/hello/toto"
    });

});