import { MainLayout } from '@/components/layouts/MainLayout';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

export default function NodesPage() {
    return (
        <MainLayout>
            <div className="p-6">
                <h1 className="text-3xl font-bold mb-6">Nodes</h1>
                <Card>
                    <CardHeader>
                        <CardTitle>Available Nodes</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p className="text-muted-foreground">Browse and manage available workflow nodes.</p>
                    </CardContent>
                </Card>
            </div>
        </MainLayout>
    );
}
