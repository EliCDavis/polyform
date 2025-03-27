import { build, serve, BuildOptions } from 'esbuild';


const buildOptsWeb: BuildOptions = {
  entryPoints: ['./website/index.ts'],
  //   inject: [],
  outfile: './generator/html/js/index.js',
  //   external: [],
  platform: 'browser',
  target: ['esNext'],
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
  servedir: './'
};

const flags = process.argv.filter(arg => /--[^=].*/.test(arg));
const enableWatch = (flags.includes('--watch'));

if (enableWatch) {

  buildOptsWeb.watch = {
    onRebuild: (error, result) => {
      if (error) { console.error('watch web development build failed:', error); }
      else { console.log('watch web development build succeeded:', result); }
    }
  };

  serve(serveOpts, {}).then((result) => {
    console.log(`serving extension from "${serveOpts.servedir}" at "http://${result.host}:${result.port}"`);
  });
}

build(buildOptsWeb).then(() => enableWatch ? console.log("watching web development build...") : null);