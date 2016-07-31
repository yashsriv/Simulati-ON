set term x11 size 1500, 1200
set xrange [-500:500]
set yrange [-500:500]
set zrange [-5:5]
splot "List.txt" with points ps 1 pt 7 title "N-Body Simulation"
pause 1
reread
