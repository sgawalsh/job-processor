import { useState } from 'react'
import { useNavigate } from 'react-router-dom';

function JobForm() {
  const [description, setDescription] = useState('');
  const [error, setError] = useState(null);
  const navigate = useNavigate();

  const submitJob = async (e) => {
    e.preventDefault();
    setError(null);

    try {
      const res = await fetch('/api/jobs', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ description }),
      });

      if (!res.ok) {
        const payload = await res.json().catch(() => ({}));
        throw new Error(payload.error || 'Failed to submit job');
      }

      const data = await res.json();

      // SUCCESS → redirect
      navigate(`/jobs/${data.id}`);
    } catch (err) {
      // FAILURE → stay here
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
    </div>
  );
}

export default JobForm;