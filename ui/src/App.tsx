import { Routes, Route } from 'react-router-dom';
import { Layout } from './components/layout';
import { HomePage } from './pages/home';
import { PreviewPage } from './pages/preview';
import { AccountsPage } from './pages/search/accounts';
import { PrincipalsPage } from './pages/search/principals';
import { ResourcesPage } from './pages/search/resources';
import { PoliciesPage } from './pages/search/policies';
import { ActionsPage } from './pages/search/actions';
import { AccessCheckPage } from './pages/simulate/access';
import { WhichPrincipalsPage } from './pages/simulate/which-principals';
import { WhichActionsPage } from './pages/simulate/which-actions';
import { WhichResourcesPage } from './pages/simulate/which-resources';
import { OverlaysPage } from './pages/overlays';
import { OverlayEditorPage } from './pages/overlays/editor';

export function App(): JSX.Element {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<HomePage />} />
        <Route path="/preview" element={<PreviewPage />} />
        <Route path="/search/accounts" element={<AccountsPage />} />
        <Route path="/search/accounts/*" element={<AccountsPage />} />
        <Route path="/search/principals" element={<PrincipalsPage />} />
        <Route path="/search/principals/*" element={<PrincipalsPage />} />
        <Route path="/search/resources" element={<ResourcesPage />} />
        <Route path="/search/resources/*" element={<ResourcesPage />} />
        <Route path="/search/policies" element={<PoliciesPage />} />
        <Route path="/search/policies/*" element={<PoliciesPage />} />
        <Route path="/search/actions" element={<ActionsPage />} />
        <Route path="/search/actions/*" element={<ActionsPage />} />
        <Route path="/simulate/access" element={<AccessCheckPage />} />
        <Route path="/simulate/which-principals" element={<WhichPrincipalsPage />} />
        <Route path="/simulate/which-actions" element={<WhichActionsPage />} />
        <Route path="/simulate/which-resources" element={<WhichResourcesPage />} />
        <Route path="/overlays" element={<OverlaysPage />} />
        <Route path="/overlays/:id/edit" element={<OverlayEditorPage />} />
        <Route path="/overlays/*" element={<OverlaysPage />} />
      </Route>
    </Routes>
  );
}
