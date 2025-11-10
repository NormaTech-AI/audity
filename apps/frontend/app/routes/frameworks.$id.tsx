import { useState, useEffect } from 'react';
import { Link, useParams, useNavigate } from 'react-router';
import { ArrowLeft, Edit, Trash2, FileText, CheckCircle2, Circle } from 'lucide-react';
import { Button } from '~/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '~/components/ui/card';
import { Badge } from '~/components/ui/badge';
import { api } from '~/api';
import type { Framework, FrameworkChecklist, FrameworkQuestion } from '~/types';
import { BackButton } from '~/components/ui/back-button';

export default function FrameworkDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [framework, setFramework] = useState<Framework | null>(null);
  const [checklist, setChecklist] = useState<FrameworkChecklist | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (id) {
      loadFramework();
      loadChecklist();
    }
  }, [id]);

  const loadFramework = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.frameworks.getById(id!);
      setFramework(response.data);
    } catch (err: any) {
      console.error('Failed to load framework:', err);
      setError(err.response?.data?.error || 'Failed to load framework');
    } finally {
      setLoading(false);
    }
  };

  const loadChecklist = async () => {
    try {
      const response = await api.frameworks.getChecklist(id!);
      const sections = response.data.map((question: FrameworkQuestion)=>{
        return question.section_title
      })
      const uniqueSections = [...new Set(sections)];
      const finalSection = uniqueSections.map((section: string) => {
        return {
          title: section,
          questions: response.data.filter((question: FrameworkQuestion) => question.section_title === section)
        }
      })
      setChecklist({sections:finalSection});
    } catch (err: any) {
      console.error('Failed to load checklist:', err);
    }
  };

  const handleDelete = async () => {
    if (!confirm('Are you sure you want to delete this framework? This action cannot be undone.')) {
      return;
    }

    try {
      await api.frameworks.delete(id!);
      navigate('/frameworks');
    } catch (err: any) {
      console.error('Failed to delete framework:', err);
      alert(err.response?.data?.error || 'Failed to delete framework');
    }
  };

  if (loading) {
    return (
      <div className="container mx-auto py-6">
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
        </div>
      </div>
    );
  }

  if (error || !framework) {
    return (
      <div className="container mx-auto py-6">
        <Card className="border-destructive">
          <CardContent className="pt-6">
            <p className="text-destructive">{error || 'Framework not found'}</p>
            <Link to="/frameworks">
              <Button className="mt-4" variant="outline">
                <ArrowLeft className="mr-2 h-4 w-4" />
                Back to Frameworks
              </Button>
            </Link>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <BackButton variant="ghost" size="icon" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">{framework.name}</h1>
            <p className="text-muted-foreground mt-2">{framework.description}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Link to={`/frameworks/${id}/edit`}>
            <Button variant="outline">
              <Edit className="mr-2 h-4 w-4" />
              Edit
            </Button>
          </Link>
          <Button variant="destructive" className='text-white/95' onClick={handleDelete}>
            <Trash2 className="mr-2 h-4 w-4" />
            Delete
          </Button>
        </div>
      </div>

      {/* Framework Info */}
      <div className="grid gap-6 md:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>Version</CardTitle>
          </CardHeader>
          <CardContent>
            <Badge variant="outline" className="text-lg">
              {framework.version}
            </Badge>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Total Questions</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-bold">{framework.question_count || 0}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Created</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-lg">
              {new Date(framework.created_at).toLocaleDateString('en-US', {
                year: 'numeric',
                month: 'long',
                day: 'numeric',
              })}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Checklist Sections */}
      {checklist && checklist.sections && checklist.sections.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Checklist Sections</CardTitle>
            <CardDescription>
              {checklist.sections.length} section{checklist.sections.length !== 1 ? 's' : ''}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {checklist.sections.map((section, sectionIndex) => (
              <div key={sectionIndex} className="space-y-4">
                <div className="border-l-4 border-primary pl-4">
                  <h3 className="text-xl font-semibold">{section.title}</h3>
                  {section.description && (
                    <p className="text-muted-foreground mt-1">{section.description}</p>
                  )}
                  <p className="text-sm text-muted-foreground mt-2">
                    {section.questions.length} question{section.questions.length !== 1 ? 's' : ''}
                  </p>
                </div>

                <div className="space-y-3 ml-4">
                  {section.questions.map((question, questionIndex) => (
                    <div
                      key={questionIndex}
                      className="flex items-start gap-3 p-4 rounded-lg border bg-card"
                    >
                      <div className="flex-1 space-y-2">
                        <div className="flex items-start justify-between gap-4">
                          <div className="flex-1">
                            <div className="flex items-center gap-2">
                              <Badge variant="outline">{question.control_id}</Badge>
                            </div>
                            <p className="mt-2 font-medium">{question.question_text}</p>
                            {question.help_text && (
                              <p className="text-sm text-muted-foreground mt-1">
                                {question.help_text}
                              </p>
                            )}
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </CardContent>
        </Card>
      )}

      {/* Empty State */}
      {(!checklist || !checklist.sections || checklist.sections.length === 0) && (
        <Card>
          <CardContent className="pt-6">
            <div className="text-center py-8">
              <FileText className="mx-auto h-12 w-12 text-muted-foreground" />
              <h3 className="mt-4 text-lg font-semibold">No checklist available</h3>
              <p className="text-muted-foreground mt-2">
                This framework doesn't have any checklist sections yet.
              </p>
              <Link to={`/frameworks/${id}/edit`}>
                <Button className="mt-4">
                  <Edit className="mr-2 h-4 w-4" />
                  Edit Framework
                </Button>
              </Link>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
