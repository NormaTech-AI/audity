import { useState, useEffect } from 'react';
import { Link, useNavigate, useParams } from 'react-router';
import { ArrowLeft, Plus, Trash2, Save } from 'lucide-react';
import { Button } from '~/components/ui/button';
import { Input } from '~/components/ui/input';
import { Label } from '~/components/ui/label';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '~/components/ui/card';
import { api } from '~/api';
import type { FrameworkQuestion } from '~/types';
import { BackButton } from '~/components/ui/back-button';

export default function EditFrameworkPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [initialLoading, setInitialLoading] = useState(true);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [version, setVersion] = useState('');
  const [questions, setQuestions] = useState<FrameworkQuestion[]>([]);
  const [evidenceInputs, setEvidenceInputs] = useState<Record<number, string>>({});

  useEffect(() => {
    if (id) {
      loadFramework();
    }
  }, [id]);

  const loadFramework = async () => {
    try {
      setInitialLoading(true);
      const [frameworkResponse, checklistResponse] = await Promise.all([
        api.frameworks.getById(id!),
        api.frameworks.getChecklist(id!),
      ]);
      
      const framework = frameworkResponse.data;
      const questions = checklistResponse.data;

      setName(framework.name);
      setDescription(framework.description);
      setVersion(framework.version);
      setQuestions(questions || []);
    } catch (err: any) {
      console.error('Failed to load framework:', err);
      alert(err.response?.data?.error || 'Failed to load framework');
      navigate('/frameworks');
    } finally {
      setInitialLoading(false);
    }
  };

  const addQuestion = () => {
    setQuestions([
      ...questions,
      {
        section_title: '',
        control_id: '',
        question_text: '',
        help_text: '',
        acceptable_evidence: [],
      },
    ]);
  };

  const removeQuestion = (questionIndex: number) => {
    setQuestions(questions.filter((_, index) => index !== questionIndex));
  };

  const updateQuestion = (
    questionIndex: number,
    field: keyof FrameworkQuestion,
    value: any
  ) => {
    const newQuestions = [...questions];
    newQuestions[questionIndex] = {
      ...newQuestions[questionIndex],
      [field]: value,
    };
    setQuestions(newQuestions);
  };

  const addEvidence = (questionIndex: number) => {
    const evidence = evidenceInputs[questionIndex] || '';
    if (!evidence.trim()) return;
    const newQuestions = [...questions];
    const currentEvidence = newQuestions[questionIndex].acceptable_evidence || [];
    newQuestions[questionIndex].acceptable_evidence = [...currentEvidence, evidence.trim()];
    setQuestions(newQuestions);
    setEvidenceInputs({ ...evidenceInputs, [questionIndex]: '' });
  };

  const removeEvidence = (questionIndex: number, evidenceIndex: number) => {
    const newQuestions = [...questions];
    const currentEvidence = newQuestions[questionIndex].acceptable_evidence || [];
    newQuestions[questionIndex].acceptable_evidence = currentEvidence.filter(
      (_, index) => index !== evidenceIndex
    );
    setQuestions(newQuestions);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!name || !description || !version) {
      setInitialLoading(false);
      return;
    }

    if (questions.length === 0) {
      setInitialLoading(false);
      return;
    }

    // Validate all questions have required fields
    const invalidQuestion = questions.find(q => !q.control_id || !q.question_text);
    if (invalidQuestion) {
      setInitialLoading(false);
      return;
    }

    try {
      setLoading(true);
      await api.frameworks.update(id!, {
        name,
        description,
        version,
        questions,
      });
      navigate(`/frameworks/${id}`);
    } catch (err: any) {
      console.error('Failed to update framework:', err);
      setInitialLoading(false);
    } finally {
      setLoading(false);
    }
  };

  if (initialLoading) {
    return (
      <div className="container mx-auto py-6">
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <BackButton variant="ghost" size="icon" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Edit Framework</h1>
          <p className="text-muted-foreground mt-2">
            Update framework details, sections and questions
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Basic Info */}
        <Card>
          <CardHeader>
            <CardTitle>Basic Information</CardTitle>
            <CardDescription>Framework details and metadata</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name">Framework Name *</Label>
              <Input
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g., SOC 2 Type II"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="description">Description *</Label>
              <Input
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Brief description of the framework"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="version">Version *</Label>
              <Input
                id="version"
                value={version}
                onChange={(e) => setVersion(e.target.value)}
                placeholder="e.g., 1.0"
                required
              />
            </div>
          </CardContent>
        </Card>

        {/* Questions */}
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-2xl font-bold">Questions</h2>
            <Button type="button" onClick={addQuestion} variant="outline">
              <Plus className="mr-2 h-4 w-4" />
              Add Question
            </Button>
          </div>

          {questions.map((question, questionIndex) => (
            <Card key={questionIndex}>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle>Question {questionIndex + 1}</CardTitle>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => removeQuestion(questionIndex)}
                  >
                    <Trash2 className="h-4 w-4 text-destructive" />
                  </Button>
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label>Section Title</Label>
                  <Input
                    value={question.section_title || ''}
                    onChange={(e) => updateQuestion(questionIndex, 'section_title', e.target.value)}
                    placeholder="e.g., Access Control"
                  />
                </div>
                <div className="space-y-2">
                  <Label>Control ID *</Label>
                  <Input
                    value={question.control_id}
                    onChange={(e) => updateQuestion(questionIndex, 'control_id', e.target.value)}
                    placeholder="e.g., 1.a.i"
                    required
                  />
                </div>
                <div className="space-y-2">
                  <Label>Question Text *</Label>
                  <Input
                    value={question.question_text}
                    onChange={(e) => updateQuestion(questionIndex, 'question_text', e.target.value)}
                    placeholder="Enter the question"
                    required
                  />
                </div>
                <div className="space-y-2">
                  <Label>Help Text</Label>
                  <Input
                    value={question.help_text || ''}
                    onChange={(e) => updateQuestion(questionIndex, 'help_text', e.target.value)}
                    placeholder="Optional guidance for answering"
                  />
                </div>
                <div className="space-y-2">
                  <Label>Acceptable Evidence</Label>
                  <div className="space-y-2">
                    {question.acceptable_evidence?.map((evidence, evidenceIndex) => (
                      <div key={evidenceIndex} className="flex items-center gap-2">
                        <Input value={evidence} className="flex-1" />
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          onClick={() => removeEvidence(questionIndex, evidenceIndex)}
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                    ))}
                    <div className="flex items-center gap-2">
                      <Input
                        placeholder="Add evidence type"
                        value={evidenceInputs[questionIndex] || ''}
                        onChange={(e) => setEvidenceInputs({ ...evidenceInputs, [questionIndex]: e.target.value })}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') {
                            e.preventDefault();
                            addEvidence(questionIndex);
                          }
                        }}
                      />
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => addEvidence(questionIndex)}
                      >
                        <Plus className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}

          {questions.length === 0 && (
            <Card>
              <CardContent className="text-center py-8 text-muted-foreground">
                No questions added yet. Click "Add Question" to get started.
              </CardContent>
            </Card>
          )}
        </div>

        {/* Actions */}
        <div className="flex items-center justify-end gap-4">
          <Link to={`/frameworks/${id}`}>
            <Button type="button" variant="outline">
              Cancel
            </Button>
          </Link>
          <Button type="submit" disabled={loading}>
            <Save className="mr-2 h-4 w-4" />
            {loading ? 'Saving...' : 'Save Changes'}
          </Button>
        </div>
      </form>
    </div>
  );
}
