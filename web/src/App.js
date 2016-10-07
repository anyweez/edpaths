import React, { Component } from 'react';
import './App.css';

import { SystemPicker } from './components/SystemPicker';
import { SystemHeading } from './components/SystemHeading';
import { RouteList } from './components/RouteList';
import { addToRoute } from './actions/route';

class App extends Component {
  add(id) {
    this.props.state.dispatch(addToRoute(id));
  }

  render() {
    let state = this.props.state.getState();
    console.log(`Stops: ${state.stops.length}`)

    return (
      <div className="App">
        <header>
          <h1>ED Paths</h1>
        </header>
        <main>
          <SystemPicker onActivate={id => this.add(id)} />
          <section className={`route-overview ${state.stops.length > 1 ? '' : 'hide'}`}>
            <SystemHeading system={state.origin} />
            <p>{state.distance}ly to</p>
            <SystemHeading system={state.destination} />
          </section>
          <RouteList stops={state.stops}></RouteList>
        </main>
      </div>
    );
  }
}

export default App;
