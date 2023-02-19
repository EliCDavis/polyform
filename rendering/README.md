# Rendering

Package for rendering meshes generated from polyform.

Straight up a 1:1 implementation based on the guide ["Ray Tracing in One Weekend" by Peter Shirley](https://raytracing.github.io/books/RayTracingInOneWeekend.html)

## Benchmarking

The demo scene from ["Ray Tracing in One Weekend"](https://raytracing.github.io/books/RayTracingInOneWeekend.html) has been put into a golang benchmark. If you try implementing optimizations, you can use this to test out what's going on.

```cmd
go test rendering/render_test.go -run BenchmarkBunnyRender -bench=BenchmarkBunnyRender -cpuprofile cpu.prof
go tool pprof -svg cpu.prof > cpu.svg
```

## Making a Video

Some examples output frames to a video that then need to get stitched together. For that I use ffmpeg:

```
ffmpeg -framerate 24 -pattern_type glob -i '*.png' -c:v libx264 -pix_fmt yuv420p -vf "pad=ceil(iw/2)*2:ceil(ih/2)*2" out.mp4
```