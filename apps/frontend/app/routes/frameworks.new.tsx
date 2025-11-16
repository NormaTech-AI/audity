import { useState } from 'react';
import { Link, useNavigate } from 'react-router';
import { Plus, Trash2, Save, Download } from 'lucide-react';
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

export default function NewFrameworkPage() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [version, setVersion] = useState('1.0');
  const [questions, setQuestions] = useState<FrameworkQuestion[]>([]);
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);

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

  const addEvidence = (questionIndex: number, evidence: string) => {
    if (!evidence.trim()) return;
    const newQuestions = [...questions];
    const currentEvidence = newQuestions[questionIndex].acceptable_evidence || [];
    newQuestions[questionIndex].acceptable_evidence = [...currentEvidence, evidence.trim()];
    setQuestions(newQuestions);
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
      setUploadError('Please fill in all required fields');
      return;
    }

    if (questions.length === 0) {
      setUploadError('Please add at least one question');
      return;
    }

    // Validate all questions have required fields
    const invalidQuestion = questions.find(q => !q.control_id || !q.question_text);
    if (invalidQuestion) {
      setUploadError('Please fill in control_id and question_text for all questions');
      return;
    }

    try {
      setLoading(true);
      await api.frameworks.create({
        name,
        description,
        version,
        questions,
      });
      navigate('/frameworks');
    } catch (err: any) {
      console.error('Failed to create framework:', err);
      setUploadError(err.response?.data?.error || 'Failed to create framework');
    } finally {
      setLoading(false);
    }
  };

  const handleExcelUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setUploading(true);
    setUploadError(null);
    try {
      const XLSX = await import('xlsx');
      const ext = file.name.split('.').pop()?.toLowerCase();
      let wb;
      if (ext === 'csv') {
        const text = await file.text();
        wb = XLSX.read(text, { type: 'string' });
      } else {
        if (file.size > 5 * 1024 * 1024) {
          throw new Error('File too large (max 5MB)');
        }
        const arrayBuffer = await file.arrayBuffer();
        wb = XLSX.read(arrayBuffer, { type: 'array' });
      }
      if (!wb.SheetNames?.length) {
        setUploadError('No sheets found in file');
        return;
      }
      const sheetName = wb.SheetNames[0];
      const sheet = wb.Sheets[sheetName];
      if (!sheet) {
        setUploadError('Could not read first sheet');
        return;
      }
      const rows = XLSX.utils.sheet_to_json<any>(sheet, { defval: '' });
      if (rows.length > 2000) {
        setUploadError('File has too many rows (max 2000)');
        return;
      }

      const parsed: FrameworkQuestion[] = rows.map((row: any, idx: number) => {
        const section = String(row['Section'] || row['section'] || '').trim();
        const control = String(row['Control ID'] || row['control_id'] || '').trim();
        const text = String(row['Question Text'] || row['question_text'] || '').trim();
        const help = String(row['Help Text'] || row['help_text'] || '').trim();
        const evidenceRaw = String(row['Acceptable Evidence'] || row['acceptable_evidence'] || '').trim();
        const evidence = evidenceRaw ? Array.from(new Set(evidenceRaw.split(',').map((s) => s.trim()).filter(Boolean))) : [];

        if (!control || !text) {
          console.warn(`Row ${idx + 2} missing control_id or question_text`);
        }

        return {
          section_title: section,
          control_id: control,
          question_text: text,
          help_text: help || undefined,
          acceptable_evidence: evidence,
        } as FrameworkQuestion;
      });

      if (parsed.length === 0) {
        setUploadError('No rows parsed. Please check template headers.');
      } else {
        setQuestions(parsed);
      }
    } catch (err: any) {
      console.error('Failed to parse spreadsheet:', err);
      setUploadError(err?.message || 'Failed to parse file');
    } finally {
      setUploading(false);
      e.target.value = '';
    }
  };

  const handleDownloadTemplate = () => {
    const headers = ['Section','Control ID','Question Text','Help Text','Acceptable Evidence'];
    const sample = [
      ['Access Control','AC-1','Establish access control policy','Provide link to policy','policy, logs'],
      ['Access Control','AC-2','User account provisioning documented','','tickets, approvals']
    ];
    const csv = [headers, ...sample]
      .map(r => r.map(v => `"${String(v).replace(/"/g,'""')}"`).join(','))
      .join('\n');
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'framework-questions-template.csv';
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <BackButton variant="ghost" size="icon" />
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Create Framework</h1>
          <p className="text-muted-foreground mt-2">
            Define a new compliance framework with questions
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

          {/* Upload & Guide */}
          <Card>
            <CardHeader>
              <CardTitle>Bulk Upload via Excel/CSV</CardTitle>
              <CardDescription>Import questions from a spreadsheet</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center gap-3">
                <input
                  type="file"
                  accept=".xlsx,.xls,.csv"
                  onChange={handleExcelUpload}
                />
                <Button type="button" variant="outline" onClick={handleDownloadTemplate}>
                  <Download className="mr-2 h-4 w-4" />
                  Download Template
                </Button>
                {uploading && (
                  <span className="text-muted-foreground" role="status" aria-live="polite">Parsing file...</span>
                )}
              </div>
              {uploadError && (
                <div className="text-destructive text-sm" role="alert">{uploadError}</div>
              )}
              <div className="text-sm text-muted-foreground">
                <p className="font-medium">Expected columns (first row as headers):</p>
                <ul className="list-disc ml-6 mt-2">
                  <li>Section (string)</li>
                  <li>Control ID (string, required)</li>
                  <li>Question Text (string, required)</li>
                  <li>Help Text (string, optional)</li>
                  <li>Acceptable Evidence (comma-separated strings)</li>
                </ul>
              </div>
            </CardContent>
          </Card>

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
                        <Input value={evidence} readOnly className="flex-1" />
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
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') {
                            e.preventDefault();
                            const input = e.currentTarget;
                            addEvidence(questionIndex, input.value);
                            input.value = '';
                          }
                        }}
                      />
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={(e) => {
                          const input = e.currentTarget.previousElementSibling as HTMLInputElement;
                          addEvidence(questionIndex, input.value);
                          input.value = '';
                        }}
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
          <Link to="/frameworks">
            <Button type="button" variant="outline">
              Cancel
            </Button>
          </Link>
          <Button type="submit" disabled={loading}>
            <Save className="mr-2 h-4 w-4" />
            {loading ? 'Creating...' : 'Create Framework'}
          </Button>
        </div>
      </form>
    </div>
  );
}
