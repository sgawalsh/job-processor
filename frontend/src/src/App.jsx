import { Routes, Route } from 'react-router-dom';
import JobForm from './JobForm';
import JobStatus from './JobStatus';
import NotFound from './NotFound';

function App() {
  return (
    <Routes>
      <Route path="/" element={<JobForm />} />
      <Route path="/jobs/:id" element={<JobStatus />} />
      <Route path="*" element={<NotFound />} />
    </Routes>
  );
}

export default App;