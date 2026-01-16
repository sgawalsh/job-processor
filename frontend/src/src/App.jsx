import { Routes, Route } from 'react-router-dom';
import JobForm from './JobForm';
import JobStatus from './JobStatus';

function App() {
  return (
    <Routes>
      <Route path="/" element={<JobForm />} />
      <Route path="/jobs/:id" element={<JobStatus />} />
    </Routes>
  );
}

export default App;