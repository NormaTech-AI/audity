import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useParams, Link } from 'react-router';
import { ArrowLeft, Save, Send, CheckCircle2, Clock, AlertCircle, User } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '~/components/ui/card';
import { Badge } from '~/components/ui/badge';
import { Button } from '~/components/ui/button';
import { Textarea } from '~/components/ui/textarea';
import { RadioGroup, RadioGroupItem } from '~/components/ui/radio-group';
import { Label } from '~/components/ui/label';
import { toast } from 'sonner';
import { api } from '~/api';
import { useAuth } from '~/contexts/AuthContext';
import type { ClientAuditQuestion, ClientSubmissionPayload } from '~/types';

export default function ClientAuditDetailPage() {
  const { auditId } = useParams();
  const { user } = useAuth();
  const queryClient = useQueryClient();

  // Local state for form data
  const [formData, setFormData] = useState<Record<string, { answer_value?: string; answer_text?: string; explanation: string }>>({});

  const { data: auditDetail, isLoading } = useQuery({
    queryKey: ['client-audit-detail', auditId],
    queryFn: async () => {
      const response = await api.clientAudit.getAuditDetail(auditId!);
      return response.data;
    },
    enabled: !!auditId,
  });

  // Save submission mutation (draft)
  const saveSubmissionMutation = useMutation({
    mutationFn: (payload: ClientSubmissionPayload) => api.clientAudit.saveSubmission(payload),
    onSuccess: () => {
      toast.success('Your answer has been saved as draft');
      queryClient.invalidateQueries({ queryKey: ['client-audit-detail', auditId] });
    },
    onError: () => {
      toast.error('Failed to save answer');
    },
  });

  // Submit answer mutation
  const submitAnswerMutation = useMutation({
    mutationFn: (submissionId: string) => api.clientAudit.submitAnswer(submissionId),
    onSuccess: () => {
      toast.success('Your answer has been submitted for review');
      queryClient.invalidateQueries({ queryKey: ['client-audit-detail', auditId] });
    },
    onError: () => {
      toast.error('Failed to submit answer');
    },
  });

  const handleSaveAnswer = (questionId: string) => {
    const data = formData[questionId];
    if (!data || !data.explanation) {
      toast.error('Explanation is required');
      return;
    }

    saveSubmissionMutation.mutate({
      question_id: questionId,
      answer_value: data.answer_value as any,
      answer_text: data.answer_text,
      explanation: data.explanation,
    });
  };

  const handleSubmitAnswer = (submissionId: string) => {
    submitAnswerMutation.mutate(submissionId);
  };

  const updateFormData = (questionId: string, field: string, value: string) => {
    setFormData(prev => ({
      ...prev,
      [questionId]: {
        ...prev[questionId],
        [field]: value,
      },
    }));
  };

  const getSubmissionStatusBadge = (status?: string) => {
    if (!status || status === 'not_started') return null;

    const statusConfig: Record<string, { label: string; icon: any; className: string }> = {
      in_progress: { label: 'Draft', icon: Clock, className: 'bg-yellow-100 text-yellow-800' },
      submitted: { label: 'Submitted', icon: Send, className: 'bg-blue-100 text-blue-800' },
      approved: { label: 'Approved', icon: CheckCircle2, className: 'bg-green-100 text-green-800' },
      rejected: { label: 'Rejected', icon: AlertCircle, className: 'bg-red-100 text-red-800' },
      referred: { label: 'Referred', icon: User, className: 'bg-purple-100 text-purple-800' },
    };

    const config = statusConfig[status];
    if (!config) return null;

    const Icon = config.icon;
    return (
      <Badge className={config.className}>
        <Icon className="h-3 w-3 mr-1" />
        {config.label}
      </Badge>
    );
  };

  // Group questions by section
  const questionsBySection = auditDetail?.questions.reduce((acc, question) => {
    if (!acc[question.section]) {
      acc[question.section] = [];
    }
    acc[question.section].push(question);
    return acc;
  }, {} as Record<string, ClientAuditQuestion[]>) || {};

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="text-center py-12">
          <p className="text-muted-foreground">Loading audit details...</p>
        </div>
      </div>
    );
  }

  if (!auditDetail) {
    return (
      <div className="space-y-6">
        <div className="text-center py-12">
          <p className="text-muted-foreground">Audit not found</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <Link to="/audit">
          <Button variant="ghost" size="icon">
            <ArrowLeft className="h-5 w-5" />
          </Button>
        </Link>
        <div className="flex-1">
          <h1 className="text-3xl font-bold tracking-tight">{auditDetail.framework_name}</h1>
          <p className="text-muted-foreground mt-1">
            Due: {new Date(auditDetail.due_date).toLocaleDateString()} • Status: {auditDetail.status}
          </p>
        </div>
      </div>

      {/* Questions by Section */}
      {Object.entries(questionsBySection).map(([section, questions]) => (
        <Card key={section}>
          <CardHeader>
            <CardTitle>{section}</CardTitle>
            <CardDescription>{questions.length} questions in this section</CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {questions.map((question) => {
              const currentData = formData[question.id] || {
                answer_value: question.answer_value,
                answer_text: question.answer_text,
                explanation: question.explanation || '',
              };

              const isReadOnly = question.submission_status === 'submitted' || 
                                question.submission_status === 'approved';

              return (
                <div key={question.id} className="border rounded-lg p-6 space-y-4">
                  {/* Question Header */}
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-2">
                        <span className="font-mono text-sm text-muted-foreground">
                          {question.question_number}
                        </span>
                        {question.is_mandatory && (
                          <Badge variant="outline" className="text-xs">Required</Badge>
                        )}
                        {question.is_assigned_to_me && (
                          <Badge variant="secondary" className="text-xs">
                            <User className="h-3 w-3 mr-1" />
                            Assigned to you
                          </Badge>
                        )}
                        {getSubmissionStatusBadge(question.submission_status)}
                      </div>
                      <h3 className="font-semibold text-lg">{question.question_text}</h3>
                      {question.help_text && (
                        <p className="text-sm text-muted-foreground mt-2">{question.help_text}</p>
                      )}
                    </div>
                  </div>

                  {/* Answer Input */}
                  <div className="space-y-4">
                    {/* Yes/No Questions */}
                    {question.question_type === 'yes_no' && (
                      <div>
                        <Label>Answer</Label>
                        <RadioGroup
                          value={currentData.answer_value}
                          onValueChange={(value) => updateFormData(question.id, 'answer_value', value)}
                          disabled={isReadOnly}
                          className="flex gap-4 mt-2"
                        >
                          <div className="flex items-center space-x-2">
                            <RadioGroupItem value="yes" id={`${question.id}-yes`} />
                            <Label htmlFor={`${question.id}-yes`}>Yes</Label>
                          </div>
                          <div className="flex items-center space-x-2">
                            <RadioGroupItem value="no" id={`${question.id}-no`} />
                            <Label htmlFor={`${question.id}-no`}>No</Label>
                          </div>
                          <div className="flex items-center space-x-2">
                            <RadioGroupItem value="na" id={`${question.id}-na`} />
                            <Label htmlFor={`${question.id}-na`}>N/A</Label>
                          </div>
                        </RadioGroup>
                      </div>
                    )}

                    {/* Text Questions */}
                    {question.question_type === 'text' && (
                      <div>
                        <Label htmlFor={`${question.id}-text`}>Answer</Label>
                        <Textarea
                          id={`${question.id}-text`}
                          value={currentData.answer_text || ''}
                          onChange={(e) => updateFormData(question.id, 'answer_text', e.target.value)}
                          disabled={isReadOnly}
                          placeholder="Enter your answer..."
                          rows={3}
                          className="mt-2"
                        />
                      </div>
                    )}

                    {/* Explanation (Required for all) */}
                    <div>
                      <Label htmlFor={`${question.id}-explanation`}>
                        Explanation <span className="text-red-500">*</span>
                      </Label>
                      <Textarea
                        id={`${question.id}-explanation`}
                        value={currentData.explanation}
                        onChange={(e) => updateFormData(question.id, 'explanation', e.target.value)}
                        disabled={isReadOnly}
                        placeholder="Provide a detailed explanation for your answer..."
                        rows={4}
                        className="mt-2"
                      />
                    </div>

                    {/* Action Buttons */}
                    {!isReadOnly && (
                      <div className="flex gap-2">
                        <Button
                          onClick={() => handleSaveAnswer(question.id)}
                          disabled={saveSubmissionMutation.isPending}
                          variant="outline"
                        >
                          <Save className="h-4 w-4 mr-2" />
                          Save Draft
                        </Button>
                        {question.submission_id && question.submission_status === 'in_progress' && (
                          <Button
                            onClick={() => handleSubmitAnswer(question.submission_id!)}
                            disabled={submitAnswerMutation.isPending}
                          >
                            <Send className="h-4 w-4 mr-2" />
                            Submit for Review
                          </Button>
                        )}
                      </div>
                    )}

                    {/* Submitted Info */}
                    {question.submitted_at && (
                      <p className="text-sm text-muted-foreground">
                        Submitted on {new Date(question.submitted_at).toLocaleString()}
                      </p>
                    )}
                  </div>
                </div>
              );
            })}
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
