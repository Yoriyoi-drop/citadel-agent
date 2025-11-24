import { MainLayout } from '@/components/layouts/MainLayout';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

export default function MarketplacePage() {
    return (
        <MainLayout>
            <div className="p-6">
                <h1 className="text-3xl font-bold mb-6">Marketplace</h1>
                <Card>
                    <CardHeader>
                        <CardTitle>Extension Marketplace</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p className="text-muted-foreground">Find and install new extensions and integrations.</p>
                    </CardContent>
                </Card>
            </div>
        </MainLayout>
    );
}
