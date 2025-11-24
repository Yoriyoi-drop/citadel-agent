import { MainLayout } from '@/components/layouts/MainLayout';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

export default function ExecutionsPage() {
    return (
        <MainLayout>
            <div className="p-6">
                <h1 className="text-3xl font-bold mb-6">Executions</h1>
                <Card>
                    <CardHeader>
                        <CardTitle>Execution History</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p className="text-muted-foreground">List of workflow executions will appear here.</p>
                    </CardContent>
                </Card>
            </div>
        </MainLayout>
    );
}
