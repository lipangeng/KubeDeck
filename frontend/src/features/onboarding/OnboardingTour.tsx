import { useState, useEffect } from 'react';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Paper from '@mui/material/Paper';
import Typography from '@mui/material/Typography';
import Step from '@mui/material/Step';
import StepContent from '@mui/material/StepContent';
import StepLabel from '@mui/material/StepLabel';
import Stepper from '@mui/material/Stepper';
import IconButton from '@mui/material/IconButton';
import CloseIcon from '@mui/icons-material/Close';

interface OnboardingStep {
  title: string;
  description: string;
  target: string;
  action?: () => void;
}

interface OnboardingTourProps {
  onComplete: () => void;
}

const steps: OnboardingStep[] = [
  {
    title: '欢迎使用 KubeDeck',
    description: 'KubeDeck 是一个现代化的 Kubernetes 控制平面，帮助您轻松管理多个集群。',
    target: 'header',
  },
  {
    title: '集群管理',
    description: '在顶部栏查看和切换当前集群。点击集群按钮可以切换到其他集群。',
    target: 'cluster-selector',
  },
  {
    title: '导航菜单',
    description: '左侧菜单显示所有可用的功能模块。点击即可导航到相应页面。',
    target: 'navigation',
  },
  {
    title: '工作负载',
    description: '在 Workloads 页面查看和管理您的 Deployment、Pod 等资源。',
    target: 'workloads',
  },
  {
    title: 'AI 助手',
    description: '右下角的 AI 助手可以帮助您执行查询和操作。随时点击咨询！',
    target: 'ai-assistant',
  },
  {
    title: '开始使用',
    description: '现在您已经了解了基本功能，开始探索 KubeDeck 吧！',
    target: 'complete',
  },
];

export function OnboardingTour({ onComplete }: OnboardingTourProps) {
  const [activeStep, setActiveStep] = useState(0);
  const [showTour, setShowTour] = useState(true);

  useEffect(() => {
    // Check if user has completed onboarding
    const completed = localStorage.getItem('onboardingCompleted');
    if (completed === 'true') {
      setShowTour(false);
      onComplete();
    }
  }, [onComplete]);

  const handleNext = () => {
    const nextStep = activeStep + 1;
    if (nextStep >= steps.length) {
      localStorage.setItem('onboardingCompleted', 'true');
      setShowTour(false);
      onComplete();
    } else {
      setActiveStep(nextStep);
    }
  };

  const handleBack = () => {
    setActiveStep((prevStep) => prevStep - 1);
  };

  const handleSkip = () => {
    localStorage.setItem('onboardingCompleted', 'true');
    setShowTour(false);
    onComplete();
  };

  if (!showTour) {
    return null;
  }

  return (
    <Paper
      elevation={6}
      sx={{
        position: 'fixed',
        top: 80,
        right: 24,
        width: 400,
        maxWidth: '90vw',
        zIndex: 1300,
        p: 2,
      }}
      role="dialog"
      aria-label="Onboarding tour"
      aria-modal="true"
    >
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h6">新手引导</Typography>
        <IconButton size="small" onClick={handleSkip} aria-label="Skip onboarding">
          <CloseIcon />
        </IconButton>
      </Box>

      <Stepper activeStep={activeStep} orientation="vertical">
        {steps.map((step, index) => (
          <Step key={step.title}>
            <StepLabel>{step.title}</StepLabel>
            <StepContent>
              <Typography variant="body2" sx={{ mb: 2 }}>
                {step.description}
              </Typography>
              <Box sx={{ mb: 2 }}>
                {index === steps.length - 1 ? (
                  <Button
                    variant="contained"
                    onClick={handleNext}
                    sx={{ mt: 1, mr: 1 }}
                  >
                    开始使用
                  </Button>
                ) : (
                  <>
                    <Button
                      variant="contained"
                      onClick={handleNext}
                      sx={{ mt: 1, mr: 1 }}
                    >
                      {index === steps.length - 2 ? '完成' : '下一步'}
                    </Button>
                    <Button
                      variant="outlined"
                      onClick={handleBack}
                      disabled={index === 0}
                      sx={{ mt: 1, mr: 1 }}
                    >
                      上一步
                    </Button>
                  </>
                )}
              </Box>
            </StepContent>
          </Step>
        ))}
      </Stepper>

      <Box sx={{ mt: 2, textAlign: 'center' }}>
        <Button size="small" onClick={handleSkip}>
          跳过引导
        </Button>
      </Box>
    </Paper>
  );
}
