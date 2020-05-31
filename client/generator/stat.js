var jStat = require('jStat');
var mean = 0 // 1s
var std = mean / 2

for (var i = 0; i < 100; ++i) {
    console.log(Math.ceil(jStat.normal.sample( mean, std )));
}

function getRandomDelay(mean) {
    var std = mean / 3;
    return Math.ceil(jStat.normal.sample( mean, std ));
}
