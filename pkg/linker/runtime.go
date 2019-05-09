package linker

const runtime = `
var cache = {};
var modules = {};

window.__modules__ = modules;

function require(name) {
  if (cache[name]) {
    return cache[name].exports;
  }
  var module = {
    name: name,
    exports: {}
  };
  modules[name](module, module.exports, require);
  cache[name] = module;
  return module.exports;
}

function chunk(path, cb) {
  var script = document.createElement('script');
  script.src = path;
  script.type = 'text/javascript';
  script.onload = function() { cb() };
  document.getElementsByTagName('head')[0].appendChild(script);
}

function start(chunks, main) {
  if (!chunks || chunks.length === 0) {
    require(main);
  }
  chunks.forEach(function(path) {
    chunk(path, function() {
      loaded++;
      if (loaded === chunks.length) {
        require(main);
      }
    });
  });
}
`
