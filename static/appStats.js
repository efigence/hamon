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
    .height(40)
    .unit("MB")
    //.min_extent([0,1])
    .colors(pallete_blues)
var chartTime = d3.horizonChart()
    .height(40)
    .unit("us")
    .min_extent([0,500])
    .colors(pallete_rainbow);
var chartCPU =
    d3.horizonChart()
        .height(40)
        .unit("s")
        .min_extent([0,4])
        .colors(pallete_greens);
var chartCPUSys =
    d3.horizonChart()
        .height(40)
        .unit("s")
        .min_extent([0,4])
        .colors(pallete_oranges);

    setInterval(function() {
        d3.json("/gcstat",function(data) {
            heapDiv
            .data([data['heap_history'].map(x => x / 1024 / 1024)])
            .each(chartBytes);
        });
         d3.json("/gcstat",function(data) {
             pauseDiv
             .data([data['pause_history']])
             .each(chartTime);
         });
        d3.json("/gcstat",function(data) {
             cpuUserDiv
             .data([data['cpu_user']])
             .each(chartCPU);
         });
        d3.json("/gcstat",function(data) {
             cpuSysDiv
             .data([data['cpu_sys']])
             .each(chartCPUSys);
         });

    },950);