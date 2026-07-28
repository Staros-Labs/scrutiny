import { appRoutes } from './app.routing';

describe('appRoutes', () => {
    it('should expose a dedicated mobile drives route', () => {
        const protectedRoute = appRoutes.find((route) => route.children?.some((child) => child.path === 'dashboard'));

        expect(protectedRoute?.children?.some((child) => child.path === 'mobile-drives')).toBeTrue();
    });

    it('should expose SMART host management without adding a mobile tab route', () => {
        const protectedRoute = appRoutes.find((route) => route.children?.some((child) => child.path === 'dashboard'));

        expect(protectedRoute?.children?.some((child) => child.path === 'hosts')).toBeTrue();
    });
});
