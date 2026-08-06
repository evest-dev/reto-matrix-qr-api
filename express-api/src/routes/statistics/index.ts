import { Router } from 'express';

import { validateBody } from '../../middleware/validateBody';
import { calculate } from './controller';
import { statisticsRequestSchema } from './schema';

const router = Router();

router.post('/', validateBody(statisticsRequestSchema), calculate);

export default router;
