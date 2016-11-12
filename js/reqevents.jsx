/* eslint-disable react/forbid-prop-types, no-console */
import 'babel-polyfill';
import React from 'react';
import ReactDOM from 'react-dom';

const locationURL = new URL(window.location);
const gomibakoKey = locationURL.pathname.match(/^\/g\/([^/]+)\/inspect/)[1];

const reqEventsPath = `/g/${gomibakoKey}/reqevents`;
const accessURL = `${new URL(window.location).origin}/g/${gomibakoKey}`;

function formatTimestamp(t) {
  return `${t.getFullYear()}-${t.getMonth() + 1}-${t.getDate()} ${t.getHours()}:${t.getMinutes()}`;
}

const Request = (props) => {
  const request = props.request;
  const headers = request.headers.map(h => (
    <tr key={h.key}><th>{h.key}</th><td>{h.value}</td></tr>
  ));

  const body = request.body ?
    (<pre>{request.body}</pre>) : (<pre className="no-body">No body</pre>);

  return (
    <div className="request">
      <div className="timestamp">{formatTimestamp(request.timestamp)}</div>
      <div className="method-url">{request.method} {request.url}</div>
      <table>
        <tbody>{headers}</tbody>
      </table>
      <div className="body-title">Body</div>
      {body}
    </div>
  );
};
Request.propTypes = {
  request: React.PropTypes.object,
};

class Requests extends React.Component {
  constructor() {
    super();
    this.state = {
      requests: [],
    };
  }
  componentDidMount() {
    const reqevents = new EventSource(reqEventsPath);
    reqevents.onmessage = (e) => {
      const r = JSON.parse(e.data);
      r.timestamp = new Date(r.timestamp * 1000);
      this.state.requests.unshift(r);
      this.setState({
        requests: this.state.requests,
      });
    };
  }
  render() {
    const requests = this.state.requests.map(r => (
      <li key={r.timestamp}><Request request={r} /></li>
    ));
    const showMsg = this.state.requests.length === 0;
    const message = showMsg ? (
      <div className="message">
        Access to {accessURL}
      </div>
    ) : (
      ''
    );
    return (
      <div>
        {message}
        <ul className="requests">
          {requests}
        </ul>
      </div>
    );
  }
}

Requests.propTypes = {
  requests: React.PropTypes.arrayOf(React.PropTypes.element),
};

ReactDOM.render(<Requests />, document.querySelector('#requests'));

