import { Button, Container, Stack, Text, Title } from '@mantine/core';
import Link from 'next/link';

export default function NotFound() {
  return (
    <Container size="sm" py="xl">
      <Stack align="center" gap="md">
        <Title order={1}>404</Title>
        <Text c="dimmed">ページが見つかりませんでした。</Text>
        <Button component={Link} href="/" variant="filled" color="blue">
          トップへ戻る
        </Button>
      </Stack>
    </Container>
  );
}
