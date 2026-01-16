import { useState } from 'react'
import './App.css'

function App() {
  const [description, setDescription] = useState('');
  const [job, setJob] = useState(null);
  const [error, setError] = useState(null);

  const submitJob = async (e) => {
    e.preventDefault();
    setError(null);

    try {
      const res = await fetch(`api/jobs`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ description }),
      });

      if (!res.ok) {
        throw new Error('Failed to submit job');
      }

      const data = await res.json();
      setJob(data);
    } catch (err) {
      setError(err.message);
    }
  };

  return (
    <div style={{ padding: 20 }}>
      <h1>Job Processor</h1>

      <form onSubmit={submitJob}>
        <input
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Job description"
        />
        <button type="submit">Submit Job</button>
      </form>

      {error && <p style={{ color: 'red' }}>{error}</p>}

      {job && (
        <div>
          <h3>Job submitted</h3>
          <p>ID: {job.id}</p>
          <p>Description: {job.description}</p>
        </div>
      )}
    </div>
  );
}

export default App;
