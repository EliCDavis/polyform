# Rendering

Package for rendering meshes generated from polyform.

Straight up a 1:1 implementation based on the guide ["Ray Tracing in One Weekend" by Peter Shirley](https://raytracing.github.io/books/RayTracingInOneWeekend.html)

## Benchmarking

The demo scene from ["Ray Tracing in One Weekend"](https://raytracing.github.io/books/RayTracingInOneWeekend.html) has been put into a golang benchmark. If you try implementing optimizations, you can use this to test out what's going on.

```cmd
go test rendering/render_test.go -run BenchmarkRender -bench=BenchmarkRender -cpuprofile cpu.prof
go tool pprof -svg cpu.prof > cpu.svg
```
