var heapDiv =  d3.select('[role="heap_history"]');
var pauseDiv =  d3.select('[role="pause_history"]');
var cpuUserDiv =  d3.select('[role="cpu_user_history"]');
var cpuSysDiv =  d3.select('[role="cpu_sys_history"]');
var pallete = [
    '#313695', '#4575b4', '#74add1', '#abd9e9',
    '#fee090', '#fdae61', '#f46d43', '#d73027'
]

var blues = ['#abd9e9','#4575b4',  '#74add1', '#26365d'];
var greens = ["#55ff55","#11aa11","#117711","#004400",];
var oranges = ["#ffdd00","#ddaa00","#aa8800","#996600"];
var reds = ["#ff5555","#aa0000","#770000","#330000"];
var rainbow = ["#4575b4","#55ff55","#ddaa00","#aa0000"];
var pallete_blues = [blues.slice().reverse(),blues].flat(2);
var pallete_greens = [greens.slice().reverse(),greens].flat(2);
var pallete_oranges = [oranges.slice().reverse(),oranges].flat(2);
var pallete_reds = [reds.slice().reverse(),reds].flat(2);
var pallete_rainbow = [rainbow.slice().reverse(),rainbow].flat(2);
var chartBytes = d3.horizonChart()
    .height(20)
    .unit("MB")
    //.min_extent([0,1])
    .colors(pallete_blues)
var chartTime = d3.horizonChart()
    .height(20)
    .unit("us")
    .min_extent([0,500])
    .colors(pallete_rainbow);
var chartCPU =
    d3.horizonChart()
        .height(20)
        .unit("s")
        .min_extent([0,4])
        .colors(pallete_greens);
var chartCPUSys =
    d3.horizonChart()
        .height(20)
        .unit("s")
        .min_extent([0,10])
        .colors(pallete_oranges);

var chartTimeMs =
    d3.horizonChart()
        .height(20)
        .unit(" ms")
        .min_extent([0,50])
        .colors(pallete_rainbow);
var chartReq = d3.horizonChart()
    .height(20)
    .unit(" rps")
    .min_extent([0,1])
    .colors(pallete_blues)
dataHash={};
dataHashUpdater={};
function dataCache(path) {
    if ( !dataHashUpdater[path] ) {
        dataHashUpdater[path] = 1
        d3.json(path, function (data) {
            dataHash[path] = data
        })
        setInterval(function () {
            d3.json(path, function (data) {
                dataHash[path] = data
            })
        },940 + (Math.random()*20))
    }
}


dataCache("/gcstat");
setInterval(function() {
    data = dataHash["/gcstat"];
        heapDiv
            .data([data['heap_history'].map(x => x / 1024 / 1024)])
            .each(chartBytes);
        pauseDiv
            .data([data['pause_history']])
            .each(chartTime);
        cpuUserDiv
            .data([data['cpu_user']])
            .each(chartCPU);
        cpuSysDiv
            .data([data['cpu_sys']])
            .each(chartCPUSys);
},950);

function runGraph(graphTemplate,path, key,key2, role) {
    var div =  d3.select('[role="' + role + '"]');
    dataCache(path);
    setInterval(function () {
        data = dataHash[path]
        div
            .data([data[key][key2]])
            .each(graphTemplate);
    },940 + (Math.random()*20));
}