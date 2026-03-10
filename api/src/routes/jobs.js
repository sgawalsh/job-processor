const express = require('express');
const { jobFailures } = require('../metrics');

function jobsRoutes({ pool }) {
    const router = express.Router();

    router.post('/', async (req, res) => {
        const { description } = req.body;
        if (!description) {
            return res.status(400).json({ error: 'Description is required' });
        }

        const client = await pool.connect();

        try {
            await client.query('BEGIN');
            const result = await client.query(
            'INSERT INTO jobs(description) VALUES($1) RETURNING public_id, description',
            [description]
            );

            await client.query('COMMIT');
            res.status(201).json(result.rows[0]);
        } catch (err) {
            await client.query('ROLLBACK');
            console.error(err);
            jobFailures.inc();
            res.status(500).json({ error: 'Database error' });
        } finally {
            client.release();
        }
    });

    router.get('/:public_id', async (req, res) => {
        res.set('Cache-Control', 'no-store');
        const { public_id } = req.params;

        try {
            const result = await pool.query(
                `SELECT id, description, status, created_at, started_at, attempts, last_error
                FROM jobs
                WHERE public_id = $1`,
                [public_id]
            );

            if (result.rowCount === 0) {
                return res.status(404).json({ error: 'Job not found' });
            }

            res.json(result.rows[0]);
        } catch (err) {
            console.error(err);
            res.status(500).json({ error: 'Database error' });
        }
    });

    return router;
}

module.exports = jobsRoutes;