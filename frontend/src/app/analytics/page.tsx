import { MainLayout } from '@/components/layouts/MainLayout';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

export default function AnalyticsPage() {
    return (
        <MainLayout>
            <div className="p-6">
                <h1 className="text-3xl font-bold mb-6">Analytics</h1>
                <Card>
                    <CardHeader>
                        <CardTitle>Performance Analytics</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p className="text-muted-foreground">View detailed statistics about your workflows.</p>
                    </CardContent>
                </Card>
            </div>
        </MainLayout>
    );
}
