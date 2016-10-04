import React from 'react';
import ReactDOM from 'react-dom';
import store from './store';

import App from './App';
// import './index.css';

const render = () => ReactDOM.render(<App state={store} />, document.getElementById('root'));

render();
store.subscribe(render);
// store.subscribe(() => console.log(store.getState()))