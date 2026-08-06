import { Router } from 'express';

import statisticsRoutes from './statistics';

const router = Router();

router.use('/statistics', statisticsRoutes);

export default router;
