import { Routes, Route } from 'react-router-dom';
import { Layout } from './components/layout';
import { HomePage } from './pages/home';
import { PreviewPage } from './pages/preview';
import { PrincipalsPage } from './pages/search/principals';
import { ResourcesPage } from './pages/search/resources';
import { PoliciesPage } from './pages/search/policies';

export function App(): JSX.Element {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<HomePage />} />
        <Route path="/preview" element={<PreviewPage />} />
        <Route path="/search/principals" element={<PrincipalsPage />} />
        <Route path="/search/principals/*" element={<PrincipalsPage />} />
        <Route path="/search/resources" element={<ResourcesPage />} />
        <Route path="/search/resources/*" element={<ResourcesPage />} />
        <Route path="/search/policies" element={<PoliciesPage />} />
        <Route path="/search/policies/*" element={<PoliciesPage />} />
      </Route>
    </Routes>
  );
}
