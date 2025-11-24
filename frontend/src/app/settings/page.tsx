import { MainLayout } from '@/components/layouts/MainLayout';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

export default function SettingsPage() {
    return (
        <MainLayout>
            <div className="p-6">
                <h1 className="text-3xl font-bold mb-6">Settings</h1>
                <Card>
                    <CardHeader>
                        <CardTitle>Application Settings</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <p className="text-muted-foreground">Configure your application preferences.</p>
                    </CardContent>
                </Card>
            </div>
        </MainLayout>
    );
}
