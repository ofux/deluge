
/// rain

var orders = {
    status: 200,
    body: [{
        plop: "toto",
        ref: 1
    }, {
        plop: "toto",
        ref: 2
    }, {
        plop: "toto",
        ref: 3
    }]
};

assert(orders.body.length == 3);
pause("10ms");

for (var i=0; i < 3; i++) {
    var newBody = orders.body[i];
    newBody.plop = "new value";

    var updatedOrder = doHTTP({
        url: "http://localhost:8080/hello/" + newBody.ref
        method: "GET",
        headers: {
            Authorization: "Bearer ojiafzojazf",
            ContentType: "application/json"
        },
        body: newBody
    });

    //assert(updatedOrder.status == 200);
}
