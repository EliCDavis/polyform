import { build, context, BuildOptions } from "esbuild";

const buildOptsWeb: BuildOptions = {
  entryPoints: ["./website/index.ts"],
  //   inject: [],
  outfile: "./generator/html/js/index.js",
  //   external: [],
  platform: "browser",
  target: ["esNext"],
  //   format: 'cjs',
  bundle: true,
  sourcemap: true,
  minify: true,
  treeShaking: true,
  plugins: [
    //     NodeModulesPolyfillPlugin(),
    //     NodeGlobalsPolyfillPlugin({
    //       process: true,
    //     }),
  ],
};

const serveOpts = {
  servedir: "./",
};

const flags = process.argv.filter((arg) => /--[^=].*/.test(arg));
const enableWatch = flags.includes("--watch");

async function startDevServer() {
  const ctx = await context(buildOptsWeb);

  if (enableWatch) {
    await ctx.watch();
    console.log("watching web development build...");

    const { hosts, port } = await ctx.serve(serveOpts);
    console.log(
      `serving extension from "${serveOpts.servedir}" at "http://${hosts[0]}:${port}"`,
    );
  } else {
    await ctx.rebuild();
    await ctx.dispose();
  }
}

startDevServer().catch((err) => {
  console.error("Build failed:", err);
  process.exit(1);
});
