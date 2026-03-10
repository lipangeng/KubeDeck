export const enMessages = {
  'app.title': 'KubeDeck',
  'app.themeLabel': 'Theme',
  'app.kernelNavigation': 'Kernel Navigation',
  'app.registeredActions': 'Registered actions',
  'app.none': 'None',
  'app.kernelMetadataSource': 'Kernel metadata source',
  'app.runAction': 'Run',
  'app.lastActionResult': 'Last action result',
  'app.cleanup.badge': 'Baseline Reset',
  'app.cleanup.title': 'Microkernel cleanup in progress',
  'app.cleanup.body':
    'Historical workflow code has been removed so the next patch can start from the accepted kernel and i18n design instead of inheriting the previous task-flow shell.',
  'remoteCapability.badge': 'Remote Capability',
  'remoteCapability.body':
    'This workflow page is rendered by the generic kernel runtime because no built-in page implementation was registered locally.',
  'homepage.title': 'Homepage',
  'homepage.description': 'The shell now renders built-in workflow pages through the kernel registry.',
  'workloads.title': 'Workloads',
  'workloads.description':
    'This is the first built-in workflow domain registered through the kernel contribution system.',
  'workloads.loading': 'Loading built-in workload data from the kernel backend.',
  'workloads.placeholder':
    'Action execution now flows through the kernel action entry instead of page-private logic.',
  'workloads.badge': 'Built-in Capability',
  'workloads.columns.name': 'Name',
  'workloads.columns.kind': 'Kind',
  'workloads.columns.namespace': 'Namespace',
  'workloads.columns.status': 'Status',
  'workloads.columns.health': 'Health',
  'actions.create': 'Create',
  'actions.apply': 'Apply',
} as const;
