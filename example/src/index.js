import React from 'react';
import ReactDOM from 'react-dom';

import("test.js").then(() => {
  console.log("HI")
})

ReactDOM.render(
  <h1>Hello, world!</h1>,
  document.getElementById('root')
);
