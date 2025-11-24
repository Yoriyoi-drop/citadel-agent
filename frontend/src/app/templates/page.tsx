import { MainLayout } from '@/components/layouts/MainLayout';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

export default function TemplatesPage() {
    return (
        <MainLayout>
            <div className="p-6">
                <h1 className="text-3xl font-bold mb-6">Templates</h1>
                <Card>
                    <CardHeader>
                        <CardTitle>Workflow Templates</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p className="text-muted-foreground">Start from a pre-built template.</p>
                    </CardContent>
                </Card>
            </div>
        </MainLayout>
    );
}
